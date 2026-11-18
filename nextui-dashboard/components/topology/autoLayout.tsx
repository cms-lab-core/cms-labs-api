'use client';
import dagre from 'dagre';
import {
  defaultStyleEdge,
  defaultStyleNodes,
  RFEdgeTopology,
  RFNodeTopology
} from '@/components/topology/objectTypes/types';

export const getLayoutElements = (nodes: RFNodeTopology[], edges: RFEdgeTopology[], direction: 'TB' | 'LR' = 'TB') => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  dagreGraph.setGraph({
    rankdir: direction,
    nodesep: 50, // Расстояние между узлами
    ranksep: 70, // Расстояние между уровнями
    marginx: 20,
    marginy: 20
  });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, {
      width: node.width || 172,
      height: node.height || 36
    });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  return {
    nodes: nodes.map((node) => {
      const { x, y } = dagreGraph.node(node.id);
      return {
        ...node,
        position: node.data.layoutPosition ?? { x, y },
        data: { ...node.data, isLayouted: true },
        style: defaultStyleNodes
      };
    }),
    edges: edges.map((edge) => ({
      ...edge,
      style: defaultStyleEdge
    }))
  };
};
