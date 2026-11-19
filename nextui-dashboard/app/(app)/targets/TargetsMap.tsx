'use client';

import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Background,
  ConnectionLineType,
  Controls,
  Edge,
  MarkerType,
  MiniMap,
  Node,
  Panel,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { getLayoutedElementsTargets, LayoutDirection } from './graph-layout';
import { TargetsNode } from '@/app/(app)/targets/TargetNodeFlow';
import { filterTargets, TargetMapFilter, TargetNodeType, TargetRelationType } from '@/app/(app)/targets/target-utils';
import { Button } from '@heroui/react';
import { ArrowDownUp, Package, Puzzle, Server, ServerCog, X } from 'lucide-react';
import { Loading } from '@/components/scroll/loader';
import { CamelCasedPropertiesDeep } from 'type-fest';
import type { UsecasesTargetListResponse } from '@/helpers/api';

const nodeTypes = {
  server: TargetsNode,
  service: TargetsNode,
  module: TargetsNode,
  virtual: TargetsNode
};

interface TargetMapProps {
  filter: TargetMapFilter;
  data?: CamelCasedPropertiesDeep<UsecasesTargetListResponse['result']>;
  isLoading?: boolean;
}

const typeColors: Record<string, { bg: string; border: string; text: string; dimmed: string }> = {
  server: {
    bg: '#dbeafe',
    border: '#3b82f6',
    text: '#1e40af',
    dimmed: '#e5e7eb'
  },
  virtual: {
    bg: '#cffafe',
    border: '#06b6d4',
    text: '#0e7490',
    dimmed: '#e5e7eb'
  },
  service: {
    bg: '#dcfce7',
    border: '#22c55e',
    text: '#166534',
    dimmed: '#e5e7eb'
  },
  module: {
    bg: '#f3e8ff',
    border: '#a855f7',
    text: '#7c3aed',
    dimmed: '#e5e7eb'
  }
};

const typeIcons: Record<string, React.ReactNode> = {
  SERVER: <Server className='h-4 w-4' />,
  VIRTUAL: <ServerCog className='h-4 w-4' />,
  SERVICE: <Package className='h-4 w-4' />,
  MODULE: <Puzzle className='h-4 w-4' />
};

interface TargetWrapperProps {
  children: React.ReactNode;
}

function TargetWrapper({ children }: TargetWrapperProps) {
  return (
    <div className='flex-1 w-full bg-gray-50 rounded-xl border border-gray-200 overflow-hidden relative'>
      {children}
    </div>
  );
}

function TargetMapInner({ filter, data, isLoading }: TargetMapProps) {
  const { fitView } = useReactFlow();
  const [layoutDirection, setLayoutDirection] = useState<LayoutDirection>('TB');
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

  const activeTargetIds = useMemo(() => {
    const filtered = filterTargets(data?.model?.targets || [], filter);
    return new Set(filtered.map((t) => t.taget.id));
  }, [data, filter]);

  const rawData = useMemo(() => {
    if (!data?.model?.targets) return { rawNodes: [], rawEdges: [] };

    const targets: TargetNodeType[] = data.model.targets;
    const relations: TargetRelationType[] = data.model.relations || [];
    const nodes: Node[] = [];
    const edges: Edge[] = [];

    targets.forEach((item) => {
      const isActive = activeTargetIds.has(item.taget.id);
      const type = item.taget.type ?? 'service'.toLowerCase();
      const colors = typeColors[item.taget.type ?? 'service'];
      const icon = typeIcons[item.taget.type ?? 'service'];

      nodes.push({
        id: item.taget.id || '',
        type: type,
        position: { x: 0, y: 0 },
        data: {
          label: item.taget.name,
          nodeId: item.taget.id,
          description: item.taget.description,
          type: type,
          icon: icon,
          isActive,
          links: item.taget.links,
          colors
        },
        style: {
          opacity: isActive ? 1 : 0.4,
          transition: 'opacity 0.3s ease'
        }
      });
    });

    relations.forEach((item) => {
      if (!item.relation) return;
      if (!item.relation.fromTargetId) return;
      if (!item.relation.toTargetId) return;
      const sourceActive = activeTargetIds.has(item.relation.fromTargetId);
      const targetActive = activeTargetIds.has(item.relation.toTargetId);
      const edgeActive = sourceActive && targetActive;

      edges.push({
        id: item.relation?.id?.toString() || '',
        source: item.relation.toTargetId,
        target: item.relation.fromTargetId,
        label: item.relation.relationType || '',
        type: 'smoothstep',
        animated: edgeActive,
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: '#6b7280'
        },
        style: {
          stroke: '#6b7280',
          strokeWidth: 2,
          strokeDasharray: '5 5',
          opacity: edgeActive ? 1 : 0.3
        },
        labelStyle: {
          fontSize: 10,
          fill: '#374151'
        },
        labelBgStyle: {
          fill: '#fff',
          fillOpacity: 0.8
        }
      });
    });

    return { rawNodes: nodes, rawEdges: edges };
  }, [data, activeTargetIds]);

  const { connectedNodeIds, connectedEdgeIds } = useMemo(() => {
    if (!selectedNodeId || rawData.rawEdges.length === 0) {
      return { connectedNodeIds: new Set<string>(), connectedEdgeIds: new Set<string>() };
    }

    const nodeIds = new Set<string>([selectedNodeId]);
    const edgeIds = new Set<string>();

    const traverseUp = (nodeId: string) => {
      rawData.rawEdges.forEach((edge) => {
        if (edge.target === nodeId && !edgeIds.has(edge.id)) {
          edgeIds.add(edge.id);
          nodeIds.add(edge.source);
          traverseUp(edge.source);
        }
      });
    };

    const traverseDown = (nodeId: string) => {
      rawData.rawEdges.forEach((edge) => {
        if (edge.source === nodeId && !edgeIds.has(edge.id)) {
          edgeIds.add(edge.id);
          nodeIds.add(edge.target);
          traverseDown(edge.target);
        }
      });
    };

    traverseUp(selectedNodeId);
    traverseDown(selectedNodeId);

    return { connectedNodeIds: nodeIds, connectedEdgeIds: edgeIds };
  }, [selectedNodeId, rawData]);

  const { nodes, edges } = useMemo(() => {
    const filteredNodes = rawData.rawNodes
      .filter((node) => !selectedNodeId || connectedNodeIds.has(node.id))
      .map((node) => ({
        ...node,
        data: {
          ...node.data,
          isHighlighted: node.id === selectedNodeId
        }
      }));

    const filteredEdges = rawData.rawEdges
      .filter((edge) => !selectedNodeId || connectedEdgeIds.has(edge.id))
      .map((edge) => ({
        ...edge,
        animated: (selectedNodeId && connectedEdgeIds.has(edge.id)) || undefined
      }));

    return { nodes: filteredNodes, edges: filteredEdges };
  }, [rawData, connectedNodeIds, connectedEdgeIds, selectedNodeId]);

  const layoutedElements = useMemo(() => {
    if (nodes.length === 0) return { nodes: [], edges: [] };
    return getLayoutedElementsTargets(nodes, edges, { direction: layoutDirection });
  }, [nodes, edges, layoutDirection]);

  const [layoutNodes, setLayoutNodes, onNodesChange] = useNodesState<Node>([]);
  const [layoutEdges, setLayoutEdges, onEdgesChange] = useEdgesState<Edge>([]);

  useEffect(() => {
    setLayoutNodes(layoutedElements.nodes);
    setLayoutEdges(layoutedElements.edges);
    if (layoutedElements.nodes.length > 0) {
      fitView({ padding: 0.2, duration: 0 });
    }
  }, [layoutedElements, setLayoutNodes, setLayoutEdges, fitView]);

  const toggleDirection = useCallback(() => {
    const newDirection = layoutDirection === 'TB' ? 'LR' : 'TB';
    setLayoutDirection(newDirection);
  }, [layoutDirection]);

  const clearSelection = useCallback(() => {
    setSelectedNodeId(null);
  }, []);

  const onNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      if (selectedNodeId === node.id) {
        clearSelection();
      } else {
        setSelectedNodeId(node.id);
      }
    },
    [selectedNodeId, clearSelection]
  );

  if (isLoading) {
    return (
      <TargetWrapper>
        <div className='w-full h-full flex items-center justify-center bg-gray-50 rounded-xl'>
          <Loading size='lg' />
        </div>
      </TargetWrapper>
    );
  }

  return (
    <TargetWrapper>
      <ReactFlow
        nodes={layoutNodes}
        edges={layoutEdges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        nodeTypes={nodeTypes}
        connectionLineType={ConnectionLineType.SmoothStep}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.1}
        maxZoom={2}
        proOptions={{ hideAttribution: true }}
      >
        <Background color='#e5e7eb' gap={20} />
        <Controls className='bg-white rounded-lg shadow-lg' />
        <MiniMap
          nodeColor={(node) => {
            const colors = typeColors[node.data?.type as string] ?? typeColors.service;
            return node.data?.isActive ? colors.border : colors.dimmed;
          }}
          className='bg-white rounded-lg shadow-lg'
          pannable
          zoomable
        />

        <Panel position='top-center' className='flex gap-2 m-2'>
          <Button variant='light' size='sm' onPress={toggleDirection} className='bg-white shadow-md hover:bg-gray-50'>
            <ArrowDownUp className='h-4 w-4 mr-2' />
            {layoutDirection === 'TB' ? 'Горизонтально' : 'Вертикально'}
          </Button>
          {selectedNodeId && (
            <Button variant='light' size='sm' onPress={clearSelection} className='bg-white shadow-md hover:bg-gray-50'>
              <X className='h-4 w-4 mr-2' />
              Сбросить
            </Button>
          )}
        </Panel>

        <Panel position='top-left' className='bg-white rounded-lg shadow-lg p-3 m-2'>
          <div className='text-xs font-medium text-gray-700 mb-2'>Легенда</div>
          <div className='space-y-1.5'>
            <div className='flex items-center gap-2'>
              <div className='w-3 h-3 rounded bg-blue-500'></div>
              <span className='text-xs text-gray-600'>Серверы</span>
            </div>
            <div className='flex items-center gap-2'>
              <div className='w-3 h-3 rounded bg-cyan-500'></div>
              <span className='text-xs text-gray-600'>Виртуальные серверы</span>
            </div>
            <div className='flex items-center gap-2'>
              <div className='w-3 h-3 rounded bg-green-500'></div>
              <span className='text-xs text-gray-600'>Сервисы</span>
            </div>
            <div className='flex items-center gap-2'>
              <div className='w-3 h-3 rounded bg-purple-500'></div>
              <span className='text-xs text-gray-600'>Модули</span>
            </div>
          </div>
        </Panel>
      </ReactFlow>
    </TargetWrapper>
  );
}

export function TargetsMap(props: TargetMapProps) {
  return (
    <ReactFlowProvider>
      <TargetMapInner {...props} />
    </ReactFlowProvider>
  );
}
