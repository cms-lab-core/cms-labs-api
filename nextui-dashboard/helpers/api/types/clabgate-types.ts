/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface QueriesDeploymentInfo {
  name?: string;
  ready_replicas?: number;
  replicas?: number;
  restarts?: number;
  status?: string;
}

export interface QueriesServiceInfo {
  cluster_ip?: string;
  external_ip?: string[];
  name?: string;
  type?: string;
}

export interface QueriesTTYDInfo {
  name?: string;
  url?: string;
}

export interface UsecasesNodeActionInputDTO {
  actions?: UsecasesNodeActionItem[];
  session_id: string;
}

export interface UsecasesNodeActionItem {
  /** enum: wipe,restart */
  action?: string;
  node?: string;
}

export interface UsecasesNodeActionOutputDTO {
  count?: number;
}

export interface UsecasesNodeActionRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "nodes.action" */
  method?: string;
  params?: UsecasesNodeActionInputDTO;
}

export interface UsecasesNodeActionResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesNodeActionOutputDTO;
}

export interface UsecasesTopologiesGetInputDTO {
  session_id: string;
}

export interface UsecasesTopologiesGetOutputDTO {
  deployments?: QueriesDeploymentInfo[];
  services?: QueriesServiceInfo[];
  topology?: string;
  ttyd?: QueriesTTYDInfo[];
}

export interface UsecasesTopologiesGetRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "topology.get" */
  method?: string;
  params?: UsecasesTopologiesGetInputDTO;
}

export interface UsecasesTopologiesGetResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesTopologiesGetOutputDTO;
}

export type QueriesSessionPhase = 'pending' | 'provisioning' | 'ready' | 'active' | 'degraded' | 'failed' | 'stopping';

export interface QueriesCheckerTaskLog {
  message?: string;
  namespace?: string;
  node?: string;
}

export interface QueriesCheckerTaskResult {
  complete?: boolean;
  description?: string;
  logs?: QueriesCheckerTaskLog[];
  title?: string;
}

export interface QueriesCheckerRun {
  check_id?: string;
  completed_at?: string;
  current_score?: number;
  error?: string;
  job_name?: string;
  logs?: string;
  max_score?: number;
  report?: string;
  result_display?: string;
  started_at?: string;
  status?: string;
  tasks?: QueriesCheckerTaskResult[];
}

export interface QueriesSessionRecord {
  attempt_id?: string;
  checker_running?: boolean;
  checker_successful?: boolean;
  checker?: QueriesCheckerRun;
  created_at?: string;
  desired_resources_accepted?: boolean;
  id?: string;
  lab_path?: string;
  message?: string;
  namespace?: string;
  owner_id?: string;
  phase?: QueriesSessionPhase;
  test_path?: string;
  task_revision?: string;
  title?: string;
  topology_ready?: boolean;
  username?: string;
  workspace_ready?: boolean;
  workspace_url?: string;
}

export interface UsecasesSessionOutputDTO {
  session?: QueriesSessionRecord;
}

export interface UsecasesSessionEnsureInputDTO {
  attempt_id?: string;
}

export interface UsecasesSessionGetInputDTO {
  session_id?: string;
}

export interface UsecasesSessionStopInputDTO {
  session_id?: string;
}

export interface UsecasesSessionListOutputDTO {
  sessions?: QueriesSessionRecord[];
}

export interface UsecasesSessionStopOutputDTO {
  stopped?: boolean;
}

export interface UsecasesSessionCheckInputDTO {
  session_id?: string;
}

export interface UsecasesSessionCheckOutputDTO {
  job_name?: string;
}

export interface UsecasesSessionOpenInputDTO {
  session_id?: string;
}

export interface UsecasesSessionOpenOutputDTO {
  url?: string;
}
