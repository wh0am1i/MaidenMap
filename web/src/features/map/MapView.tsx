import { MapContainer, TileLayer } from "react-leaflet";
import type { ReactNode } from "react";
import type { TileProvider } from "@/lib/tile-providers";
import "leaflet/dist/leaflet.css";

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
      scrollWheelZoom
      className="w-full h-full"
      key={provider.id} // force remount when provider changes
    >
      <TileLayer url={provider.url} attribution={provider.attribution} maxZoom={provider.maxZoom} />
      {children}
    </MapContainer>
  );
}
