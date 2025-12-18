import { apiGet, apiPost, apiDelete, apiFetch } from "./apiClient";
import { B4Config } from "@models/config";
import {
  Capture,
  GeoFileInfo,
  GeodatDownloadResult,
  GeodatSource,
  ResetResponse,
  RestartResponse,
  SystemInfo,
  UpdateResponse,
} from "@b4.settings";

// Config API
export const configApi = {
  get: () => apiGet<B4Config>("/api/config"),
  save: (config: B4Config) =>
    apiFetch<void>("/api/config", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(config),
    }),
  reset: () => apiPost<ResetResponse>("/api/config/reset"),
};

// Capture API
export const captureApi = {
  list: () => apiGet<Capture[]>("/api/capture/list"),
  probe: (domain: string, protocol: string) =>
    apiPost<void>("/api/capture/probe", { domain, protocol }),
  delete: (protocol: string, domain: string) =>
    apiDelete(`/api/capture/delete?protocol=${protocol}&domain=${domain}`),
  clear: () => apiPost<void>("/api/capture/clear"),
  downloadUrl: (filepath: string) =>
    `/api/capture/download?file=${encodeURIComponent(filepath)}`,
};

// Geodat API
export const geodatApi = {
  sources: () => apiGet<GeodatSource[]>("/api/geodat/sources"),
  info: (path: string) =>
    apiGet<GeoFileInfo>(`/api/geodat/info?path=${encodeURIComponent(path)}`),
  download: (geositeUrl: string, geoipUrl: string, destPath: string) =>
    apiPost<GeodatDownloadResult>("/api/geodat/download", {
      geosite_url: geositeUrl,
      geoip_url: geoipUrl,
      destination_path: destPath,
    }),
};

// System API
export const systemApi = {
  info: () => apiGet<SystemInfo>("/api/system/info"),
  restart: () => apiPost<RestartResponse>("/api/system/restart"),
  update: (version?: string) =>
    apiPost<UpdateResponse>("/api/system/update", { version }),
  version: () => apiGet<unknown>("/api/version"),
};
