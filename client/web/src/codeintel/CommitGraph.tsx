import React, { useEffect } from 'react'

import { SigmaContainer, useLoadGraph, ControlsContainer, ZoomControl } from '@react-sigma/core'
import { LayoutForceAtlas2Control, useLayoutForceAtlas2 } from '@react-sigma/layout-forceatlas2'
import { LayoutNoverlapControl, useLayoutNoverlap } from '@react-sigma/layout-noverlap'

import '@react-sigma/core/lib/react-sigma.min.css'

import Graph from 'graphology'

const LoadGraph = (): null => {
    const loadGraph = useLoadGraph()
    // const { assign } = useLayoutForceAtlas2()
    const { assign } = useLayoutNoverlap()

    useEffect(() => {
        const graph = new Graph({ multi: true, allowSelfLoops: true, type: 'directed' })
        for (let iter = 0; iter < 100; iter++) {
            graph.addNode(iter, { x: Math.random(), y: Math.random(), label: `My ${iter}th node` })
        }
        for (let iter = 0; iter < 200; iter++) {
            const source = Math.floor(Math.random() * 100)
            const target = Math.floor(Math.random() * 100)
            if (graph.hasDirectedEdge(source, target)) {
                continue
            }
            graph.addDirectedEdge(source, target, {
                label: `My ${iter}th edge`,
            })
        }
        loadGraph(graph)
        assign()
    })
    return null
}

export const CommitGraph: React.FunctionComponent<{}> = () => (
    <SigmaContainer
        style={{ height: '500px' }}
        settings={
            {
                // nodeProgramClasses: { image: getNodeProgramImage() },
                // defaultNodeType: 'image',
                // defaultEdgeType: 'arrow',
                // labelDensity: 0.07,
                // labelGridCellSize: 60,
                // labelRenderedSizeThreshold: 15,
                // labelFont: 'Lato, sans-serif',
                // zIndex: true,
            }
        }
    >
        <LoadGraph />
        <ControlsContainer position="bottom-right">
            <ZoomControl />
            {/* <LayoutForceAtlas2Control /> */}
            <LayoutNoverlapControl />
        </ControlsContainer>
    </SigmaContainer>
)
