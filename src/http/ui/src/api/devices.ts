import { apiGet, apiPut, apiDelete } from "./apiClient";
import { DevicesResponse } from "@b4.devices";

export const devicesApi = {
  list: () => apiGet<DevicesResponse>("/api/devices"),
  setAlias: (mac: string, alias: string) =>
    apiPut<void>(`/api/devices/${encodeURIComponent(mac)}/alias`, { alias }),
  resetAlias: (mac: string) =>
    apiDelete(`/api/devices/${encodeURIComponent(mac)}/alias`),
};
