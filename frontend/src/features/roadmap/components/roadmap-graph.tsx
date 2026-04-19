"use client";

import { useMemo, useEffect } from "react";
import {
  ReactFlow,
  Controls,
  Background,
  Node,
  Edge,
  MarkerType,
  useNodesState,
  useEdgesState,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import type { Roadmap, RoadmapCourse } from "@/types/roadmap";
import dagre from "dagre";

interface RoadmapGraphProps {
  roadmap: Roadmap | null;
}

const nodeWidth = 260;
const nodeHeight = 100;

// Helper to normalize course codes (remove spaces for matching)
const normCode = (c: string) => c.replace(/\s+/g, "").toUpperCase();

const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = "LR") => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  
  dagreGraph.setGraph({ 
    rankdir: direction, 
    nodesep: 80, 
    ranksep: 200,
    marginx: 50,
    marginy: 50
  });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { 
      width: nodeWidth, 
      height: nodeHeight,
      // Pass the academic term sequence as a 'rank' hint to Dagre
      rank: node.data.rank
    });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  nodes.forEach((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    node.targetPosition = direction === "LR" ? ("left" as any) : ("top" as any);
    node.sourcePosition = direction === "LR" ? ("right" as any) : ("bottom" as any);

    // Subtle randomization for a more organic 'web' feel
    const saltX = (Math.random() - 0.5) * 20;
    const saltY = (Math.random() - 0.5) * 40;

    node.position = {
      x: nodeWithPosition.x - nodeWidth / 2 + saltX,
      y: nodeWithPosition.y - nodeHeight / 2 + saltY,
    };
  });

  return { nodes, edges };
};

export function RoadmapGraph({ roadmap }: RoadmapGraphProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

  useEffect(() => {
    if (!roadmap) return;

    const initialNodes: Node[] = [];
    const initialEdges: Edge[] = [];
    
    // Maps normalized code -> display code (for edge creation)
    const normalizedToDisplay = new Map<string, string>();
    const courseToTermRank = new Map<string, number>();

    // Pass 1: Collect nodes and term-based ranks
    roadmap.terms.forEach((term) => {
      const rank = term.sequence; // Higher sequence = further right
      term.requirements.forEach(req => {
        if (req.course) {
          normalizedToDisplay.set(normCode(req.course.code), req.course.code);
          courseToTermRank.set(req.course.code, rank);
        }
        if (req.group) {
          normalizedToDisplay.set(normCode(req.group.code), req.group.code);
          courseToTermRank.set(req.group.code, rank);
        }
      });
    });

    const discoveredPrereqs = new Set<string>();
    // Matches ACTSC 231 or ACTSC231
    const courseRegex = /[A-Z]{2,4}\s?\d{3}[A-Z]?/g;

    // Pass 2: Discover prerequisites
    roadmap.terms.forEach(term => {
      term.requirements.forEach(req => {
        if (req.course) {
          const text = (req.course.prerequisiteMessage || "") + " " + (req.course.description || "");
          let match;
          while ((match = courseRegex.exec(text)) !== null) {
            const code = match[0];
            const nc = normCode(code);
            if (!normalizedToDisplay.has(nc)) {
              discoveredPrereqs.add(code);
            }
          }
        }
      });
    });

    // Node factory
    const createNode = (id: string, label: string, subtitle?: string, status: string = "NOT_STARTED", isPrereq: boolean = false, rank: number = 0) => {
      const statusColors: Record<string, string> = {
        COMPLETED: "#daf1ea",
        IN_PROGRESS: "#fde8cd",
        PLANNED: "#ffffff",
        NOT_STARTED: isPrereq ? "#f8fafc" : "#ffffff",
      };
      
      const borderColors: Record<string, string> = {
        COMPLETED: "#2dd4bf",
        IN_PROGRESS: "#fbbf24",
        PLANNED: "#102336",
        NOT_STARTED: isPrereq ? "#94a3b8" : "#e2e8f0",
      };

      return {
        id,
        position: { x: 0, y: 0 },
        data: {
          rank,
          label: (
            <div className="flex flex-col gap-1 text-left select-none">
              <div className="flex items-center justify-between">
                <strong className="font-display text-base text-ink">{id}</strong>
                {isPrereq && <span className="text-[9px] font-bold uppercase tracking-widest text-ink/40">Prereq</span>}
              </div>
              <p className="truncate text-xs text-ink/60">{label}</p>
              {subtitle && <p className="text-[10px] italic text-ink/40">{subtitle}</p>}
            </div>
          ),
        },
        style: {
          background: statusColors[status] || "#ffffff",
          border: isPrereq ? `2px dashed ${borderColors[status]}` : `2px solid ${borderColors[status]}`,
          borderRadius: "1rem",
          padding: "16px",
          width: nodeWidth,
          boxShadow: isPrereq ? "none" : "0 10px 15px -3px rgb(0 0 0 / 0.05)",
          opacity: isPrereq ? 0.7 : 1,
        },
      };
    };

    // Add nodes
    roadmap.terms.forEach((term) => {
      term.requirements.forEach((req) => {
        if (req.course) {
          initialNodes.push(createNode(req.course.code, req.course.title, `Planned: ${term.label}`, req.course.status, false, term.sequence));
        } else if (req.group) {
          initialNodes.push(createNode(req.group.code, req.group.title, "Elective Group", "NOT_STARTED", false, term.sequence));
        }
      });
    });

    discoveredPrereqs.forEach(code => {
      initialNodes.push(createNode(code, "External Prerequisite", "External Credit", "NOT_STARTED", true, 0));
      normalizedToDisplay.set(normCode(code), code);
    });

    // Build edges
    const edgeSet = new Set<string>();
    roadmap.terms.forEach((term) => {
      term.requirements.forEach((req) => {
        if (req.course) {
          const text = (req.course.prerequisiteMessage || "") + " " + (req.course.description || "");
          let match;
          while ((match = courseRegex.exec(text)) !== null) {
            const mentionedCode = match[0];
            const nc = normCode(mentionedCode);
            const targetCode = normalizedToDisplay.get(nc);
            
            if (targetCode && targetCode !== req.course.code) {
              const edgeId = `e-${targetCode}-${req.course.code}`;
              if (!edgeSet.has(edgeId)) {
                edgeSet.add(edgeId);
                initialEdges.push({
                  id: edgeId,
                  source: targetCode,
                  target: req.course.code,
                  animated: req.course.status === "IN_PROGRESS",
                  style: { stroke: "#94a3b8", strokeWidth: 2 },
                  markerEnd: {
                    type: MarkerType.ArrowClosed,
                    color: "#94a3b8",
                  },
                });
              }
            }
          }
        }
      });
    });

    const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
      initialNodes,
      initialEdges,
      "LR"
    );

    setNodes(layoutedNodes);
    setEdges(layoutedEdges);
  }, [roadmap, setNodes, setEdges]);

  if (!roadmap) {
    return (
      <div className="flex h-full min-h-[400px] items-center justify-center rounded-[2rem] border border-white/60 bg-white/80 p-12 text-center shadow-panel backdrop-blur">
        <h2 className="font-display text-2xl text-ink">Select a program to begin</h2>
      </div>
    );
  }

  return (
    <div style={{ width: "100%", height: "800px" }} className="rounded-[2.4rem] border border-white/60 bg-white/55 shadow-panel backdrop-blur overflow-hidden">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
        minZoom={0.05}
        maxZoom={1.5}
      >
        <Background gap={24} size={1} color="#cbd5e1" />
        <Controls />
      </ReactFlow>
    </div>
  );
}
