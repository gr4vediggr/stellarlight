import React, { useState, useRef, useEffect, useCallback, useMemo } from 'react';
import type { Galaxy, StarSystem } from '../types/galaxy';


// Viewport and rendering types
type Viewport = {
  x: number;
  y: number;
  width: number;
  height: number;
  zoom: number;
};

type RenderableObject = {
  id: string;
  type: string;
  x: number;
  y: number;
  render: (ctx: CanvasRenderingContext2D, viewport: Viewport, isSelected: boolean, isHovered: boolean) => void;
  getBounds: () => { x: number; y: number; width: number; height: number };
  onClick?: () => void;
  data: any;
};

export class PointObject implements RenderableObject {
  x: number; 
  y: number; 
  id: string;
  type: string = 'point';
  data: any = null;

  constructor(x: number, y: number) {
    this.x = x;
    this.y = y;
    this.id = 'point-' + x + '-' + y;
    this.data = { x, y };
  }

  render(ctx: CanvasRenderingContext2D, viewport: Viewport, isSelected: boolean, isHovered: boolean) {
    const screenX = (this.x - viewport.x) * viewport.zoom;
    const screenY = (this.y - viewport.y) * viewport.zoom;
    const size = 5 * viewport.zoom;

    ctx.fillStyle = isSelected ? '#fbbf24' : isHovered ? '#f59e0b' : '#888';
    ctx.beginPath();
    ctx.arc(screenX, screenY, size, 0, Math.PI * 2);
    ctx.fill();

  if (viewport.zoom > 1) {
        ctx.fillStyle = '#fff';
        ctx.font = `${Math.min(12 * viewport.zoom, 16)}px Arial`;
        ctx.textAlign = 'center';
        ctx.fillText(this.x + ',' + this.y, screenX, screenY + size + 15);
    }
  }
  getBounds() {
    const size = 5; // Generous bounds for clicking
    return {
      x: this.x - size,
      y: this.y - size,
      width: size * 2,
      height: size * 2
    };
  }


}

function transformToScreenSpace(x: number, y: number, viewport: Viewport) {
  const screenX = (x - viewport.x) * viewport.zoom + viewport.width / 2;
  const screenY = (y - viewport.y) * viewport.zoom + viewport.height / 2;
  return { x: screenX, y: screenY };
}

// Renderable Star System class
class StarSystemRenderable implements RenderableObject {
  id: string;
  type = 'starSystem';
  x: number;
  y: number;
  data: StarSystem;

  constructor(starSystem: StarSystem) {
    this.id = starSystem.id;
    this.x = starSystem.locationX;
    this.y = starSystem.locationY;
    this.data = starSystem;
  }

  render(ctx: CanvasRenderingContext2D, viewport: Viewport, isSelected: boolean, isHovered: boolean) {
    const { x: screenX, y: screenY } = transformToScreenSpace(this.x, this.y, viewport);
    const size = 6;
    

    // Don't render if too small or outside screen
    if (size < 2) return;
    
    // Determine color based on ownership
    let color = '#888';
    if (this.data.ownerId) {
      const colors = ['#4f46e5', '#dc2626', '#059669', '#d97706', '#7c3aed'];
      const index = this.data.ownerId.charCodeAt(0) % colors.length;
      color = colors[index];
    }

    // Draw star system
    ctx.fillStyle = isSelected ? '#fbbf24' : isHovered ? '#f59e0b' : color;
    ctx.beginPath();
    ctx.arc(screenX, screenY, size, 0, Math.PI * 2);
    ctx.fill();

    // Draw selection ring
    if (isSelected) {
      ctx.strokeStyle = '#fbbf24';
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.arc(screenX, screenY, size + 4, 0, Math.PI * 2);
      ctx.stroke();
    }

    // Draw name if zoomed in enough
    if (viewport.zoom > 1) {
      ctx.fillStyle = '#fff';
      ctx.font = `${Math.min(12 * viewport.zoom, 16)}px Arial`;
      ctx.textAlign = 'center';
      ctx.fillText(this.data.name, screenX, screenY + size + 15);
    }

    // Draw connections
    if (viewport.zoom > 0.5) {
      ctx.strokeStyle = '#444';
      ctx.lineWidth = 1;
     
    }
  }

  getBounds() {
    const size = 20; // Generous bounds for clicking
    return {
      x: this.x - size,
      y: this.y - size,
      width: size * 2,
      height: size * 2
    };
  }
}

// Spatial partitioning for performance
class QuadTree {
  private bounds: { x: number; y: number; width: number; height: number };
  private objects: RenderableObject[] = [];
  private children: QuadTree[] = [];
  private maxObjects = 10;
  private maxLevels = 5;
  private level = 0;

  constructor(bounds: { x: number; y: number; width: number; height: number }, level = 0) {
    this.bounds = bounds;
    this.level = level;
  }

  clear() {
    this.objects = [];
    this.children = [];
  }

  split() {
    const subWidth = this.bounds.width / 2;
    const subHeight = this.bounds.height / 2;
    const x = this.bounds.x;
    const y = this.bounds.y;

    // Order: NW, NE, SW, SE
    this.children = [
      new QuadTree({ x, y, width: subWidth, height: subHeight }, this.level + 1), // NW
      new QuadTree({ x: x + subWidth, y, width: subWidth, height: subHeight }, this.level + 1), // NE
      new QuadTree({ x, y: y + subHeight, width: subWidth, height: subHeight }, this.level + 1), // SW
      new QuadTree({ x: x + subWidth, y: y + subHeight, width: subWidth, height: subHeight }, this.level + 1) // SE
    ];
  }

  getIndices(bounds: { x: number; y: number; width: number; height: number }): number[] {
    // Returns all child indices that the bounds intersect
    const indices: number[] = [];
    const verticalMidpoint = this.bounds.x + this.bounds.width / 2;
    const horizontalMidpoint = this.bounds.y + this.bounds.height / 2;

    // NW
    if (bounds.x < verticalMidpoint && bounds.y < horizontalMidpoint) indices.push(0);
    // NE
    if (bounds.x + bounds.width > verticalMidpoint && bounds.y < horizontalMidpoint) indices.push(1);
    // SW
    if (bounds.x < verticalMidpoint && bounds.y + bounds.height > horizontalMidpoint) indices.push(2);
    // SE
    if (bounds.x + bounds.width > verticalMidpoint && bounds.y + bounds.height > horizontalMidpoint) indices.push(3);

    return indices;
  }

  insert(obj: RenderableObject) {
    if (this.children.length > 0) {
      const indices = this.getIndices(obj.getBounds());
      if (indices.length > 0) {
        indices.forEach(idx => this.children[idx].insert(obj));
        return;
      }
    }

    this.objects.push(obj);

    if (this.objects.length > this.maxObjects && this.level < this.maxLevels) {
      if (this.children.length === 0) {
        this.split();
      }

      let i = 0;
      while (i < this.objects.length) {
        const indices = this.getIndices(this.objects[i].getBounds());
        if (indices.length > 0) {
          indices.forEach(idx => this.children[idx].insert(this.objects[i]));
          this.objects.splice(i, 1);
        } else {
          i++;
        }
      }
    }
  }

  retrieve(bounds: { x: number; y: number; width: number; height: number }): RenderableObject[] {
    let objects = [...this.objects];

    if (this.children.length > 0) {
      const indices = this.getIndices(bounds);
      if (indices.length > 0) {
        indices.forEach(idx => {
          objects = objects.concat(this.children[idx].retrieve(bounds));
        });
      } else {
        // If not in any child, check all children
        this.children.forEach(child => {
          objects = objects.concat(child.retrieve(bounds));
        });
      }
    }

    // Remove duplicates by id
    const seen = new Set<string>();
    const uniqueObjects: RenderableObject[] = [];
    for (const obj of objects) {
      if (!seen.has(obj.id)) {
        seen.add(obj.id);
        uniqueObjects.push(obj);
      }
    }
    return uniqueObjects;
  }
}

// Main Galaxy Map Component
export const GalaxyMap: React.FC<{ galaxy: Galaxy }> = ({ galaxy }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  // Use effect to set initial width/height after mount
  const [viewport, setViewport] = useState<Viewport>({
    x: 0,
    y: 0,
    zoom: 1,
    width: 900,
    height: 900
  });
  const [selectedObject, setSelectedObject] = useState<RenderableObject | null>(null);
  const [hoveredObject, setHoveredObject] = useState<RenderableObject | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [showSystemView, setShowSystemView] = useState<StarSystem | null>(null);
  const [needsRender, setNeedsRender] = useState(true);
  const lastRenderState = useRef({ viewport, selectedId: '', hoveredId: '' });
  const { objectMap, quadTree } = useRenderableData(galaxy);

 

  // Check if render is needed
  const shouldRender = useCallback(() => {
    const currentState = {
      viewport: { ...viewport },
      selectedId: selectedObject?.id || '',
      hoveredId: hoveredObject?.id || ''
    };
    
    const lastState = lastRenderState.current;
    const viewportChanged = 
      Math.abs(currentState.viewport.x - lastState.viewport.x) > 0.1 ||
      Math.abs(currentState.viewport.y - lastState.viewport.y) > 0.1 ||
      Math.abs(currentState.viewport.zoom - lastState.viewport.zoom) > 0.01;
    
    const selectionChanged = 
      currentState.selectedId !== lastState.selectedId ||
      currentState.hoveredId !== lastState.hoveredId;
    
    if (viewportChanged || selectionChanged || needsRender) {
      lastRenderState.current = currentState;
      setNeedsRender(false);
      return true;
    }
    
    return false;
  }, [viewport, selectedObject, hoveredObject, needsRender]);

const render = useCallback(() => {
  if (!shouldRender()) return;

  const canvas = canvasRef.current;
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  if (!ctx) return;

  ctx.fillStyle = '#000';
  ctx.fillRect(0, 0, canvas.width, canvas.height);
  viewport.width = canvas.width;
  viewport.height = canvas.height;

  // Visible bounds
  const visibleBounds = {
    x: viewport.x - canvas.width / (2 * viewport.zoom),
    y: viewport.y - canvas.height / (2 * viewport.zoom),
    width: canvas.width / viewport.zoom,
    height: canvas.height / viewport.zoom
  };

  // Retrieve visible objects
  const visibleObjects = quadTree.retrieve(visibleBounds);

  // Precompute screen positions
  const screenPos = new Map<string, { x: number; y: number }>();
  for (const obj of visibleObjects) {
    const pos = transformToScreenSpace(obj.x, obj.y, viewport);
    screenPos.set(obj.id, pos);
  }

  // Batch draw connections
  ctx.strokeStyle = '#333';
  ctx.lineWidth = 1;
  ctx.beginPath();
  for (const obj of visibleObjects) {
    if (obj.type === 'starSystem') {
      const system = obj.data as StarSystem;
      const p1 = screenPos.get(obj.id);
      if (!p1) continue;

      for (const connectedId of system.connectedSystems) {
        const connected = objectMap.get(connectedId);
        if (!connected) continue;
        const p2 = screenPos.get(connected.id);
        if (!p2) continue;

        ctx.moveTo(p1.x, p1.y);
        ctx.lineTo(p2.x, p2.y);
      }
    }
  }
  ctx.stroke();

  // Draw objects
  for (const obj of visibleObjects) {
    const isSelected = selectedObject?.id === obj.id;
    const isHovered = hoveredObject?.id === obj.id;
    obj.render(ctx, viewport, isSelected, isHovered);
  }

  // Overlay info
  ctx.fillStyle = '#fff';
  ctx.font = '14px Arial';
  ctx.textAlign = 'left';
  ctx.fillText(`Zoom: ${viewport.zoom.toFixed(2)}x`, 10, 25);
  ctx.fillText(`Systems: ${galaxy.starSystems.length}`, 10, 45);
  if (selectedObject) {
    ctx.fillText(`Selected: ${selectedObject.data.name}`, 10, 65);
  }
}, [shouldRender, viewport, selectedObject, hoveredObject, galaxy, quadTree, objectMap]);

  // Mouse event handlers
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    const rect = canvasRef.current?.getBoundingClientRect();
    if (!rect) return;

    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    
    setIsDragging(true);
    setDragStart({ x, y });

    // Convert screen coordinates to world coordinates
    const worldX = viewport.x + (x - rect.width / 2) / viewport.zoom;
    const worldY = viewport.y + (y - rect.height / 2) / viewport.zoom;
    
    console.log('Click at screen:', x, y, 'world:', worldX, worldY);
    
    // Find objects near the click
    const tolerance = 30; // Fixed screen-space tolerance
    const worldTolerance = tolerance / viewport.zoom;
    const clickBounds = { 
      x: worldX - worldTolerance, 
      y: worldY - worldTolerance, 
      width: worldTolerance * 2, 
      height: worldTolerance * 2 
    };
    
    const nearbyObjects = quadTree.retrieve(clickBounds);
    console.log('Nearby objects:', nearbyObjects.length);
    
    // Find the closest clickable object
    let closestObject: RenderableObject | null = null;
    let closestDistance = Infinity;
    
    nearbyObjects.forEach((obj: RenderableObject) => {
      const dx = obj.x - worldX;
      const dy = obj.y - worldY;
      const distance = Math.sqrt(dx * dx + dy * dy);
      
      
      if (distance <= worldTolerance && distance < closestDistance) {
        closestObject = obj;
        closestDistance = distance;
      }
    });

    if (closestObject) {
      let renderableClosestObject = closestObject as RenderableObject;
      console.log('Selected:', renderableClosestObject.data.name);
      setSelectedObject(closestObject);
      if (renderableClosestObject.type === 'starSystem' && e.detail === 2) { // Double click
        setShowSystemView(renderableClosestObject.data as StarSystem);
        setIsDragging(false); // Stop dragging on double click
      }
    } else {
      console.log('No object selected');
      setSelectedObject(null);
    }
  }, [viewport, quadTree]);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    const rect = canvasRef.current?.getBoundingClientRect();
    if (!rect) return;

    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    if (isDragging) {
      const dx = (x - dragStart.x) / viewport.zoom;
      const dy = (y - dragStart.y) / viewport.zoom;
      
      setViewport(prev => ({
        ...prev,
        x: prev.x - dx,
        y: prev.y - dy
      }));
      
      setDragStart({ x, y });
    } else {
      // Handle hover - convert screen coordinates to world coordinates
      const worldX = viewport.x + (x - rect.width / 2) / viewport.zoom;
      const worldY = viewport.y + (y - rect.height / 2) / viewport.zoom;
      
      // Find objects near the mouse
      const tolerance = 30; // Fixed screen-space tolerance
      const worldTolerance = tolerance / viewport.zoom;
      const hoverBounds = { 
        x: worldX - worldTolerance, 
        y: worldY - worldTolerance, 
        width: worldTolerance * 2, 
        height: worldTolerance * 2 
      };
      
      const nearbyObjects = quadTree.retrieve(hoverBounds);
      
      // Find the closest hoverable object
      let closestObject: RenderableObject | null = null;
      let closestDistance = Infinity;
      
      nearbyObjects.forEach(obj => {
        const dx = obj.x - worldX;
        const dy = obj.y - worldY;
        const distance = Math.sqrt(dx * dx + dy * dy);
        
        if (distance <= worldTolerance && distance < closestDistance) {
          closestObject = obj;
          closestDistance = distance;
        }
      });

      if (hoveredObject !== closestObject) {
        setHoveredObject(closestObject);
      }
    }
  }, [isDragging, dragStart, viewport, quadTree, hoveredObject]);

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);

  const handleWheel = useCallback((e: React.WheelEvent) => {
    const zoomFactor = e.deltaY > 0 ? 0.9 : 1.1;
    setViewport(prev => ({
      ...prev,
      zoom: Math.max(0.1, Math.min(5, prev.zoom * zoomFactor))
    }));
  }, []);

  // Resize canvas to fit container and update viewport width/height
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const resizeCanvas = () => {
      const container = canvas.parentElement;
      if (container) {
        canvas.width = container.clientWidth;
        canvas.height = container.clientHeight;
        setViewport(prev => ({
          ...prev,
          width: container.clientWidth,
          height: container.clientHeight
        }));
      }
    };

    resizeCanvas();
    window.addEventListener('resize', resizeCanvas);
    return () => window.removeEventListener('resize', resizeCanvas);
  }, []);

  // Render loop - only renders when needed
  useEffect(() => {
    let animationId: number;
    
    const animate = () => {
      render();
      animationId = requestAnimationFrame(animate);
    };
    
    animationId = requestAnimationFrame(animate);
    return () => cancelAnimationFrame(animationId);
  }, [render]);

  return (
    <div className="w-full h-screen bg-gray-900 relative overflow-hidden">
      <canvas
        ref={canvasRef}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onWheel={handleWheel}
        className="w-full h-full cursor-grab active:cursor-grabbing"
      />
      
      {/* Controls */}
      <div className="absolute top-4 right-4 bg-gray-800 p-4 rounded-lg text-white">
        <div className="flex flex-col gap-2">
          <button 
            onClick={() => {
              setViewport(prev => ({ ...prev, zoom: Math.min(5, prev.zoom * 1.5) }));
            }}
            className="px-3 py-1 bg-blue-600 rounded hover:bg-blue-700"
          >
            Zoom In
          </button>
          <button 
            onClick={() => {
              setViewport(prev => ({ ...prev, zoom: Math.max(0.1, prev.zoom / 1.5) }));
            }}
            className="px-3 py-1 bg-blue-600 rounded hover:bg-blue-700"
          >
            Zoom Out
          </button>
          <button 
            onClick={() => {
              setViewport({ x: 0, y: 0, zoom: 1 , width: canvasRef.current?.width || 900, height: canvasRef.current?.height || 900 });
            }}
            className="px-3 py-1 bg-green-600 rounded hover:bg-green-700"
          >
            Reset View
          </button>
        </div>
      </div>

      {/* Object Info Panel */}
      {selectedObject && (
        <div className="absolute bottom-4 left-4 bg-gray-800 p-4 rounded-lg text-white max-w-sm">
          <h3 className="text-lg font-bold mb-2">{selectedObject.data.name}</h3>
          {selectedObject.type === 'starSystem' && (
            <div>
              <p>Type: Star System</p>
              <p>Stars: {selectedObject.data.stars.length}</p>
              <p>Owner: {selectedObject.data.ownerId || 'Uncontrolled'}</p>
              <button 
                onClick={() => setShowSystemView(selectedObject.data)}
                className="mt-2 px-3 py-1 bg-purple-600 rounded hover:bg-purple-700"
              >
                View System Details
              </button>
            </div>
          )}
        </div>
      )}

      {/* System Detail Modal */}
      {showSystemView && SystemView(showSystemView, setShowSystemView)}
    </div>
  );
};

function SystemView(system: StarSystem, returnToMap: (starSystem: StarSystem | null) => void) {
  return (
    <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
      <div className="bg-gray-800 p-6 rounded-lg text-white max-w-2xl max-h-96 overflow-y-auto pointer-events-auto shadow-2xl">
        <h2 className="text-2xl font-bold mb-4">{system.name}</h2>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <h3 className="text-lg font-semibold mb-2">Stars ({system.stars.length})</h3>
            {system.stars.map(star => (
              <div key={star.id} className="mb-2">
                <p className="font-medium">{star.name}</p>
                <p className="text-sm text-gray-300">Type: {star.type}, Size: {star.size}</p>
                <p className="text-sm text-gray-300">Planets: {star.planets.length}</p>
              </div>
            ))}
          </div>
          <div>
            <h3 className="text-lg font-semibold mb-2">System Info</h3>
            <p>Position: ({system.locationX}, {system.locationY})</p>
            <p>Owner: {system.ownerId || 'Uncontrolled'}</p>
            <p>Connected Systems: {system.connectedSystems.length}</p>
            {system.colonies && (
              <p>Colonies: {system.colonies.length}</p>
            )}
          </div>
        </div>
        <button
          onClick={() => returnToMap(null)}
          className="mt-4 px-4 py-2 bg-red-600 rounded hover:bg-red-700"
        >
          Close
        </button>
      </div>
    </div>
  );
}

// Demo data
export const createDemoGalaxy = (): Galaxy => {
  const starSystems: StarSystem[] = [];
  
  for (let i = 0; i < 100; i++) {
    const system: StarSystem = {
      id: `system-${i}`,
      name: `System ${i + 1}`,
      ownerId: Math.random() > 0.7 ? `player-${Math.floor(Math.random() * 5)}` : undefined,
      locationX: (Math.random() - 0.5) * 2000,
      locationY: (Math.random() - 0.5) * 2000,
      connectedSystems: [],
      stars: [
        {
          id: `star-${i}-0`,
          name: `Star ${i + 1}A`,
          type: ['Red Dwarf', 'Yellow Dwarf', 'Blue Giant', 'White Dwarf'][Math.floor(Math.random() * 4)],
          size: Math.random() * 3 + 1,
          planets: Array.from({ length: Math.floor(Math.random() * 8) }, (_, j) => ({
            id: `planet-${i}-${j}`,
            name: `Planet ${i + 1}.${j + 1}`,
            type: ['Rocky', 'Gas Giant', 'Ice', 'Desert'][Math.floor(Math.random() * 4)],
            size: Math.random() * 2 + 0.5,
            orbitRadius: (j + 1) * 50,
            angle: Math.random() * Math.PI * 2
          }))
        }
      ]
    };
    
    starSystems.push(system);
  }

  // Add some connections
  starSystems.forEach(system => {
    const nearbyCount = Math.floor(Math.random() * 4);
    for (let i = 0; i < nearbyCount; i++) {
      const nearby = starSystems[Math.floor(Math.random() * starSystems.length)];
      if (nearby.id !== system.id && !system.connectedSystems.includes(nearby.id)) {
        system.connectedSystems.push(nearby.id);
      }
    }
  });

  return {
    id: 'demo-galaxy',
    name: 'Demo Galaxy',
    starSystems
  };
};

// Keep this near your GalaxyMap component
function useRenderableData(galaxy: Galaxy) {
  // 1. Create renderable objects once per galaxy change
  const renderableObjects = useMemo(() => {
    const objs: RenderableObject[] = galaxy.starSystems.map(
      system => new StarSystemRenderable(system)
    );



    return objs;
  }, [galaxy]);

  // 2. Build an ID → object map for fast lookups
  const objectMap = useMemo(() => {
    const map = new Map<string, RenderableObject>();
    renderableObjects.forEach(o => map.set(o.id, o));
    return map;
  }, [renderableObjects]);

  // 3. Build the QuadTree once
  const quadTree = useMemo(() => {
    const bounds = { x: -5000, y: -5000, width: 10000, height: 10000 };
    const tree = new QuadTree(bounds);
    renderableObjects.forEach(obj => tree.insert(obj));
    return tree;
  }, [renderableObjects]);

  return { renderableObjects, objectMap, quadTree };
}
