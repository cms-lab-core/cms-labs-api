package tasks

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/app/queries"
	"github.com/cms-lab-core/cms-labs-api/shared/connection"
	"github.com/cms-lab-core/cms-labs-api/shared/proxmox_client"
	"github.com/rs/zerolog"
	"gorm.io/datatypes"
)

// ProxmoxSyncUC — use case для синхронизации инфраструктуры Proxmox.
type ProxmoxSyncUC struct {
	Debug                 bool
	ProxmoxSyncConfig     *connection.ProxmoxSyncConfig
	TargetQueries         *queries.TargetQueries
	TargetRelationQueries *queries.TargetRelationQueries
	*zerolog.Logger
}

// Execute запускает процесс синхронизации Proxmox.
func (u *ProxmoxSyncUC) Execute() error {
	ctx := context.Background()

	for _, server := range u.ProxmoxSyncConfig.Servers {
		client, err := proxmox_client.NewProxmoxAPI(u.Debug, server.Url, server.Token)
		if err != nil {
			u.Logger.Error().Err(err).Msg("Не удалось создать клиент Proxmox")
			continue
		}
		u.Logger.Info().Str("url", server.Url).Msg("Обработка сервера")

		nodes, err := client.GetNodes(ctx)
		if err != nil {
			u.Logger.Error().Err(err).Str("url", server.Url).Msg("Не удалось получить список узлов")
			continue
		}

		for _, node := range nodes {
			u.Logger.Info().Str("node", node.Node).Msg("Обработка узла")
			if err := u.processNode(ctx, client, node); err != nil {
				u.Logger.Error().Err(err).Str("node", node.Node).Msg("Не удалось обработать узел")
			}
		}
	}
	u.Logger.Info().Msg("Синхронизация Proxmox успешно завершена")
	return nil
}

// processNode обрабатывает один узел Proxmox (физический сервер) и его ВМ.
func (u *ProxmoxSyncUC) processNode(
	ctx context.Context,
	client proxmox_client.Interface,
	node proxmox_client.Node,
) error {
	nodeName := node.Node
	serverID := "server-" + nodeName

	// Формируем управляемые ссылки для сервера (CPU, RAM, диск)
	var serverLinks []models.TargetLink
	if node.MaxCPU > 0 {
		serverLinks = append(serverLinks, models.TargetLink{Value: strconv.Itoa(node.MaxCPU), Type: "cpu"})
	}
	if node.MaxMem > 0 {
		gb := float64(node.MaxMem) / 1024 / 1024 / 1024
		serverLinks = append(serverLinks, models.TargetLink{Value: fmt.Sprintf("%.2f Гб", gb), Type: "ram"})
	}
	if node.MaxDisk > 0 {
		gb := float64(node.MaxDisk) / 1024 / 1024 / 1024
		serverLinks = append(serverLinks, models.TargetLink{Value: fmt.Sprintf("%.2f Гб", gb), Type: "disk"})
	}

	desc := fmt.Sprintf("Физический сервер Proxmox (узел %s)", nodeName)
	// Для сервера теги не синхронизируем, поэтому передаем nil.
	if err := u.upsertTarget(serverID, models.TargetTypeServer, nodeName, &desc, serverLinks, nil); err != nil {
		return fmt.Errorf("не удалось выполнить upsert для цели сервера: %w", err)
	}

	// Получаем все QEMU ВМ на этом узле
	vms, err := client.GetQEMUVMs(ctx, nodeName)
	if err != nil {
		return fmt.Errorf("не удалось получить список QEMU ВМ на узле %s: %w", nodeName, err)
	}

	for _, vm := range vms {
		if err := u.processVM(ctx, client, nodeName, vm, serverID); err != nil {
			u.Logger.Error().Err(err).Int("vmid", vm.VMID).Msg("Не удалось обработать ВМ")
		}
	}
	return nil
}

// processVM обрабатывает одну виртуальную машину.
func (u *ProxmoxSyncUC) processVM(ctx context.Context, client proxmox_client.Interface, nodeName string, vm proxmox_client.QEMUVM, serverID string) error {
	vmIDFull := fmt.Sprintf("virtual-server-%d", vm.VMID)
	vmName := vm.Name
	if vmName == "" {
		vmName = fmt.Sprintf("%d", vm.VMID)
	}

	// Пытаемся получить сетевые интерфейсы от гостевого агента
	var ipv4, ipv6 string
	ifaces, err := client.GetVMNetworkInterfaces(ctx, nodeName, vm.VMID)
	if err != nil {
		u.Logger.Debug().Err(err).Int("vmid", vm.VMID).Msg("Не удалось получить сетевые интерфейсы (гостевой агент может быть не запущен)")
	} else {
		for _, iface := range ifaces {
			if iface.Name == "lo" {
				continue
			}
			for _, ip := range iface.IPAddresses {
				if ip.IPAddressType == "ipv4" && ipv4 == "" {
					ipv4 = ip.IPAddress
				} else if ip.IPAddressType == "ipv6" && ipv6 == "" {
					ipv6 = ip.IPAddress
				}
			}
		}
	}

	// Формируем управляемые ссылки для ВМ
	var managedLinks []models.TargetLink
	if vm.CPUs > 0 {
		managedLinks = append(managedLinks, models.TargetLink{Value: strconv.Itoa(vm.CPUs), Type: "cpu"})
	}
	if vm.MaxMem > 0 {
		gb := float64(vm.MaxMem) / 1024 / 1024 / 1024
		managedLinks = append(managedLinks, models.TargetLink{Value: fmt.Sprintf("%.2f Гб", gb), Type: "ram"})
	}
	if vm.MaxDisk > 0 {
		gb := float64(vm.MaxDisk) / 1024 / 1024 / 1024
		managedLinks = append(managedLinks, models.TargetLink{Value: fmt.Sprintf("%.2f Гб", gb), Type: "disk"})
	}
	if ipv4 != "" {
		managedLinks = append(managedLinks, models.TargetLink{Value: ipv4, Type: "ip"})
	}
	if ipv6 != "" {
		managedLinks = append(managedLinks, models.TargetLink{Value: ipv6, Type: "ip"})
	}

	desc := fmt.Sprintf("Виртуальная машина QEMU (ID: %d)", vm.VMID)
	// Передаем теги, полученные из конфигурации ВМ
	if err := u.upsertTarget(vmIDFull, models.TargetTypeVirtual, vmName, &desc, managedLinks, vm.Tags); err != nil {
		return fmt.Errorf("не удалось выполнить upsert для цели ВМ: %w", err)
	}

	// Создаем или обновляем связь между ВМ и физическим сервером
	relation := &models.TargetRelation{
		FromTargetID: vmIDFull,
		ToTargetID:   serverID,
		RelationType: "server",
	}
	if err := u.TargetRelationQueries.Upsert(relation); err != nil {
		return fmt.Errorf("не удалось выполнить upsert для связи: %w", err)
	}

	return nil
}

// upsertTarget создает или обновляет цель, объединяя существующие пользовательские ссылки и удаляя тег "deleted".
// Параметр newTags может быть nil, если теги не должны обновляться (для серверов).
func (u *ProxmoxSyncUC) upsertTarget(
	id string,
	targetType string,
	name string,
	description *string,
	managedLinks []models.TargetLink,
	newTags []string,
) error {
	internalTags := datatypes.NewJSONSlice[string](newTags)
	internalLinks := datatypes.NewJSONSlice[models.TargetLink](managedLinks)
	existing, err := u.TargetQueries.Get(id)
	if err != nil && !errors.Is(err, queries.TargetNotFoundError) {
		return err
	}
	if existing.ID == "" {
		existing.ID = id
		existing.CreatedAt = time.Now().UTC()
	}
	existing.Type = targetType
	existing.Name = name
	existing.Description = description
	existing.InternalTags = &internalTags
	existing.InternalLinks = &internalLinks
	existing.SynchronizedAt = time.Now().UTC()
	existing.UpdatedAt = time.Now().UTC()
	return u.TargetQueries.Upsert(&existing)
}
