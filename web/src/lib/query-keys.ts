export const qk = {
  grid: (code: string) => ["grid", code] as const,
  gridBatch: (codes: readonly string[]) => ["grid-batch", [...codes].sort().join(",")] as const,
  health: () => ["health"] as const,
};

export type GridQueryKey = ReturnType<typeof qk.grid>;
export type GridBatchQueryKey = ReturnType<typeof qk.gridBatch>;
export type HealthQueryKey = ReturnType<typeof qk.health>;
