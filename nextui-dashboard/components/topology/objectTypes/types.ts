import { Edge as RFEdge, Node as RFNode } from '@xyflow/react';
import { CSSProperties } from 'react';

export interface Edge {
  data?: EdgeData;
  id?: string;
  source?: string;
  target?: string;
  type?: string; // 'default'
}

export interface EdgeData {
  source?: EdgeDataItem;
  target?: EdgeDataItem;
}

export interface EdgeDataItem {
  id?: string;
  label?: string;
  name?: string;
  type?: string;
}

export interface Node {
  /** Explicit canvas coordinates from labels.flow_x / labels.flow_y. */
  layoutPosition?: { x: number; y: number };
  data?: NodeDataItem;
  /** Icon enum:cloud,router,server,switch,desktop */
  icon?: string; // .labels.flow_icon или kind
  id?: string; // название ноды
  label?: string; // .labels.flow_label или название ноды
  type?: string; // 'default'
}

export interface NodeDataItem {
  /** Icon enum:cloud,router,server,switch,desktop */
  icon?: string; // .labels.flow_icon или kind
  id?: string; // название ноды
  kind?: string; // .kind
  label?: string; // .name
  type?: string; // 'default'
  serviceExternalIp?: string[];
  serviceClusterIp?: string;
  serviceHealthy?: boolean;
  shellUrl?: string;
}

// @ts-ignore
export type RFNodeTopology = RFNode<Node>;

// @ts-ignore
export type RFEdgeTopology = RFEdge<Edge>;

export const defaultStyleNodes: CSSProperties = {
  width: 'auto',
  height: 'auto',
  background: 'none',
  border: 'none',
  padding: 0,
  margin: 0,
  cursor: 'pointer',
  zIndex: 10,
  boxShadow: 'none'
};

export const defaultStyleEdge: CSSProperties = {
  stroke: '#0066AAFF',
  strokeWidth: 2
};
