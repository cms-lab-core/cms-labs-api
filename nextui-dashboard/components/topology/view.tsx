'use client';
import React, { type MouseEvent as ReactMouseEvent, useCallback, useEffect, useRef, useState } from 'react';
import { Background, Controls, ReactFlow, useReactFlow } from '@xyflow/react';
import { edgeTypes, nodeTypes } from './objectTypes';
import { getLayoutElements } from './autoLayout';
import { ErrorModal } from '@/components/pages/auth/error';
import { Button } from '@heroui/react';
import { TerminalActionFunc } from '@/components/topology/terminal/service/context';
import useLanguageBrowser from '@/helpers/locale';
import useThemeBrowser from '@/components/navbar/useTheme';
import { useQueryTopologyGet } from '@/helpers/queries/topology/use-query-topology-get';
import { useParams } from 'react-router-dom';
import { parseTopology } from '@/components/topology/objectTypes/parse';
import AuthLoadingWrapper from '@/components/pages/auth/loader';
import { RFContextMenu, RFContextMenuProps } from '@/components/topology/menu';
import { RFNodeTopology } from '@/components/topology/objectTypes/types';

interface TopologyFlowVisualizationProps {
  dispatch?: TerminalActionFunc;
  highlightedNodeId?: string;
  active?: boolean;
}

export const TopologyFlowVisualization = ({
  dispatch,
  highlightedNodeId,
  active = true
}: TopologyFlowVisualizationProps) => {
  const { theme } = useThemeBrowser();
  const [menu, setMenu] = useState<RFContextMenuProps>({});
  const onNodeContextMenu = useCallback(
    (event: ReactMouseEvent, node: RFNodeTopology) => {
      // Prevent native context menu from showing
      event.preventDefault();

      // Calculate position of the context menu. We want to make sure it
      // doesn't get positioned off-screen.
      // @ts-ignore
      const pane = ref?.current?.getBoundingClientRect();
      setMenu({
        id: node.id,
        top: event.clientY < pane.height - 200 && event.clientY,
        left: event.clientX < pane.width - 200 && event.clientX,
        right: event.clientX >= pane.width - 200 && pane.width - event.clientX,
        bottom: event.clientY >= pane.height - 200 && pane.height - event.clientY
      });
    },
    [setMenu]
  );

  // Close the context menu if it's open whenever the window is clicked.
  const onPaneClick = useCallback(() => setMenu({}), [setMenu]);
  const {
    locale: {
      Topology: { Connect }
    }
  } = useLanguageBrowser();
  const sessionId = useParams().sessionId || '';
  const ref = useRef(null);
  const queryTopology = useQueryTopologyGet({ sessionId });
  const { fitView } = useReactFlow();
  useEffect(() => {
    if (!active || !queryTopology.data) return;
    const frame = window.requestAnimationFrame(() => {
      fitView({ padding: 0.16, duration: 250 });
    });
    return () => window.cancelAnimationFrame(frame);
  }, [active, fitView, queryTopology.data]);
  useEffect(() => {
    // @ts-ignore
    window.document.title = `Lab ${sessionId}`;
    return () => {
      window.document.title = 'CMS LABS';
    };
  }, [sessionId]);
  if (queryTopology.isLoading) return <AuthLoadingWrapper />;
  if (queryTopology.error || queryTopology.data === undefined) {
    // @ts-ignore
    const errMsg = queryTopology?.error?.data?.message || Connect.Error;
    return (
      <ErrorModal title={Connect.ErrorModalViewTitle} description={errMsg}>
        <Button onPress={() => queryTopology.refetch()} href='#' variant='light' color='primary'>
          {Connect.ErrorModalRetry}
        </Button>
      </ErrorModal>
    );
  }
  const parsedTopology = parseTopology(queryTopology.data);
  const object = getLayoutElements(parsedTopology.nodes, parsedTopology.edges, parsedTopology.direction);
  const nodes = object.nodes.map((node) => ({
    ...node,
    className: node.id === highlightedNodeId ? 'lab-node-highlight' : undefined
  }));

  return (
    <ReactFlow
      ref={ref}
      colorMode={theme}
      nodes={nodes}
      // @ts-ignore
      edges={object.edges}
      nodeTypes={nodeTypes}
      edgeTypes={edgeTypes}
      onNodeClick={onNodeContextMenu}
      onPaneClick={onPaneClick}
      fitView={true}
    >
      <RFContextMenu {...menu} onClick={onPaneClick} dispatch={dispatch} />
      <Background />
      <Controls />
    </ReactFlow>
  );
};
