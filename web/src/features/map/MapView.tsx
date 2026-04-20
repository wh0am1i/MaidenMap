import { MapContainer, TileLayer } from "react-leaflet";
import type { LatLngBoundsLiteral } from "leaflet";
import type { ReactNode } from "react";
import type { TileProvider } from "@/lib/tile-providers";
import "leaflet/dist/leaflet.css";

// Keep pans + zooms within a single world copy. Without these clamps the map
// lets you zoom out to see Earth tiled horizontally, which is confusing for a
// grid-locator tool.
const WORLD_BOUNDS: LatLngBoundsLiteral = [
  [-85, -180],
  [85, 180],
];
const MIN_ZOOM = 2;

export function MapView({
  provider,
  center,
  zoom,
  children,
}: {
  provider: TileProvider;
  center: [number, number];
  zoom: number;
  children?: ReactNode;
}) {
  return (
    <MapContainer
      center={center}
      zoom={zoom}
      minZoom={MIN_ZOOM}
      maxZoom={provider.maxZoom}
      maxBounds={WORLD_BOUNDS}
      maxBoundsViscosity={1}
      scrollWheelZoom
      worldCopyJump={false}
      className="w-full h-full"
      key={provider.id} // force remount when provider changes
    >
      <TileLayer
        url={provider.url}
        attribution={provider.attribution}
        maxZoom={provider.maxZoom}
        noWrap
        bounds={WORLD_BOUNDS}
      />
      {children}
    </MapContainer>
  );
}
