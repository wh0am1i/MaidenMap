import { CircleMarker, Tooltip, useMap } from "react-leaflet";
import { useEffect } from "react";

export interface BatchPin {
  grid: string;
  seq: number;
  lat: number;
  lon: number;
}

export function BatchOverlay({
  pins,
  activeGrid,
  fitOnChange,
}: {
  pins: BatchPin[];
  activeGrid: string | null;
  fitOnChange?: boolean;
}) {
  const map = useMap();

  useEffect(() => {
    if (!fitOnChange || pins.length === 0) return;
    const lats = pins.map((p) => p.lat);
    const lons = pins.map((p) => p.lon);
    map.fitBounds(
      [
        [Math.min(...lats), Math.min(...lons)],
        [Math.max(...lats), Math.max(...lons)],
      ],
      { padding: [40, 40], maxZoom: 4 },
    );
  }, [pins, fitOnChange, map]);

  return (
    <>
      {pins.map((p) => {
        const active = p.grid === activeGrid;
        return (
          <CircleMarker
            key={p.grid}
            center={[p.lat, p.lon]}
            radius={active ? 10 : 7}
            pathOptions={{
              color: "rgb(var(--ham))",
              fillColor: active ? "#fff" : "rgb(var(--ham))",
              fillOpacity: 1,
              weight: active ? 4 : 2,
            }}
          >
            <Tooltip direction="top" permanent offset={[0, -8]} className="text-[10px]">
              {p.seq}
            </Tooltip>
          </CircleMarker>
        );
      })}
    </>
  );
}
