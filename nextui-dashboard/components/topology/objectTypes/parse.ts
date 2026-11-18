import * as yaml from 'js-yaml';
import type { Edge, EdgeData, Node, NodeDataItem, RFEdgeTopology, RFNodeTopology } from './types';
import type { CamelCasedPropertiesDeep } from 'type-fest';
import type {
  QueriesDeploymentInfo,
  QueriesServiceInfo,
  QueriesTTYDInfo,
  UsecasesTopologiesGetOutputDTO
} from '@/helpers/api';

type Response = CamelCasedPropertiesDeep<UsecasesTopologiesGetOutputDTO>;
type Service = CamelCasedPropertiesDeep<QueriesServiceInfo>;
type TTYD = CamelCasedPropertiesDeep<QueriesTTYDInfo>;
type Deployment = CamelCasedPropertiesDeep<QueriesDeploymentInfo>;

// ─── Вспомогательные типы ───────────────────────────────────────────────────

interface Endpoint {
  node: string;
  iface: string;
}

interface TopologyYaml {
  topology?: {
    nodes?: Record<
      string,
      {
        labels?: Record<string, string>;
        kind?: string;
      }
    >;
    links?: Array<{ endpoints?: string[] }>;
    defaults?: {
      labels?: {
        flow_direction?: string;
      };
    };
  };
}

export interface ParsedTopology {
  nodes: RFNodeTopology[];
  edges: RFEdgeTopology[];
  direction: 'TB' | 'LR';
}

// ─── Helpers ────────────────────────────────────────────────────────────────

// Labels are strings in containerlab; tolerate unquoted YAML numbers as well.
function parseLayoutPosition(labels?: Record<string, string>): { x: number; y: number } | undefined {
  const coordinate = (value: unknown): number | undefined => {
    if (typeof value !== 'string' && typeof value !== 'number') return undefined;
    const text = String(value).trim();
    if (!/^[+-]?(?:\d+(?:\.\d*)?|\.\d+)$/.test(text)) return undefined;
    const number = Number(text);
    return Number.isFinite(number) ? number : undefined;
  };
  const x = coordinate(labels?.flow_x);
  const y = coordinate(labels?.flow_y);
  return x === undefined || y === undefined ? undefined : { x, y };
}

function parseEndpoint(ep: string): Endpoint {
  const [node, iface = ''] = ep.split(':');
  return { node, iface };
}

function findNode(nodes: RFNodeTopology[], name: string): RFNodeTopology | undefined {
  return nodes.find((n) => n.id === name);
}

function findDeployment(deployments: Deployment[], name: string): Deployment | undefined {
  return deployments.find((n) => n.name === name);
}

function findService(services: Service[], name: string): Service | undefined {
  return services.find((n) => n.name === name);
}

function findTTYD(ttyds: TTYD[], name: string): TTYD | undefined {
  return ttyds.find((n) => (n.name || '').startsWith(`${name}-`) || n.name === name);
}

// ─── Основная функция ───────────────────────────────────────────────────────

export function parseTopology(response: Response): ParsedTopology {
  // 2. Парсим вложенный YAML containerlab
  const topology = yaml.load(response.topology || '') as TopologyYaml;

  const flowDirection = (topology?.topology?.defaults?.labels?.flow_direction ?? '').toUpperCase();

  const direction = flowDirection === 'LR' ? 'LR' : 'TB';

  const nodes: RFNodeTopology[] = [];
  const edges: RFEdgeTopology[] = [];

  // 3. Обрабатываем узлы
  for (const [nodeName, nodeData] of Object.entries(topology?.topology?.nodes ?? {})) {
    const label = nodeName || nodeData.kind || '';
    const icon = nodeData.labels?.['flow_icon'] || nodeData.kind || '';
    const layoutPosition = parseLayoutPosition(nodeData.labels);

    const deployment = findDeployment(response.deployments || [], nodeName);
    const service = findService(response.services || [], nodeName);
    const ttyd = findTTYD(response.ttyd || [], nodeName);

    const dataItem: NodeDataItem = {
      id: nodeName,
      label,
      type: 'default',
      kind: nodeData.kind,
      icon,
      serviceExternalIp: service?.externalIp,
      serviceClusterIp: service?.clusterIp,
      serviceHealthy: deployment?.status === 'Ready',
      shellUrl: ttyd?.url
    };

    const nodePayload: Node = {
      id: nodeName,
      label,
      type: 'default',
      icon,
      layoutPosition,
      data: dataItem
    };

    nodes.push({
      id: nodeName,
      type: 'default',
      position: layoutPosition ?? { x: 0, y: 0 },
      data: nodePayload
    });
  }

  // 4. Обрабатываем связи
  for (let i = 0; i < (topology?.topology?.links?.length ?? 0); i++) {
    const link = topology!.topology!.links![i];

    if (!link.endpoints || link.endpoints.length < 2) continue;

    const src = parseEndpoint(link.endpoints[0]);
    const tgt = parseEndpoint(link.endpoints[1]);

    const sourceNode = findNode(nodes, src.node);
    const targetNode = findNode(nodes, tgt.node);

    if (!sourceNode && !targetNode) continue;

    const edgeData: EdgeData = {};

    if (sourceNode) {
      edgeData.source = {
        id: src.node,
        name: src.iface,
        type: sourceNode.data?.type,
        label: sourceNode.data?.label
      };
    }

    if (targetNode) {
      edgeData.target = {
        id: tgt.node,
        name: tgt.iface,
        type: targetNode.data?.type,
        label: targetNode.data?.label
      };
    }

    const edgeId = `edge-${i}`;

    const edgePayload: Edge = {
      id: edgeId,
      type: 'default',
      source: src.node,
      target: tgt.node,
      data: edgeData
    };

    edges.push({
      id: edgeId,
      type: 'default',
      source: src.node,
      target: tgt.node,
      data: edgePayload
    });
  }

  return { nodes, edges, direction };
}
