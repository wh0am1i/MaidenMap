import { CircleMarker, Rectangle, Tooltip, useMap } from "react-leaflet";
import { useEffect } from "react";
import type { ParsedGrid } from "@/lib/maidenhead";
import { parseMaidenhead } from "@/lib/maidenhead";

export function GridOverlay({ grid, fitOnMount }: { grid: ParsedGrid; fitOnMount?: boolean }) {
  const map = useMap();
  const outer = grid.length >= 6 ? parseMaidenhead(grid.normalized.slice(0, 4)) : null;
  const outerOk = outer && outer.kind === "ok" ? outer : null;

  useEffect(() => {
    if (!fitOnMount) return;
    const b = grid.bounds;
    map.fitBounds(
      [
        [b.latMin, b.lonMin],
        [b.latMax, b.lonMax],
      ],
      { padding: [32, 32] },
    );
  }, [grid, fitOnMount, map]);

  return (
    <>
      {outerOk && (
        <Rectangle
          bounds={[
            [outerOk.bounds.latMin, outerOk.bounds.lonMin],
            [outerOk.bounds.latMax, outerOk.bounds.lonMax],
          ]}
          pathOptions={{ color: "rgb(var(--ham))", weight: 1, dashArray: "4 4", fillOpacity: 0.04 }}
        >
          <Tooltip direction="top" permanent sticky className="text-[10px]">
            {outerOk.normalized}
          </Tooltip>
        </Rectangle>
      )}
      <Rectangle
        bounds={[
          [grid.bounds.latMin, grid.bounds.lonMin],
          [grid.bounds.latMax, grid.bounds.lonMax],
        ]}
        pathOptions={{ color: "rgb(var(--ham))", weight: 2, fillOpacity: 0.15 }}
      >
        <Tooltip direction="top" permanent className="text-[10px]">
          {grid.normalized}
        </Tooltip>
      </Rectangle>
      <CircleMarker
        center={[grid.lat, grid.lon]}
        radius={5}
        pathOptions={{ color: "rgb(var(--ham))", fillColor: "#fff", fillOpacity: 1, weight: 3 }}
      />
    </>
  );
}
