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

export interface AuthJWK {
  alg?: string;
  e?: string;
  kid?: string;
  kty?: string;
  n?: string;
  use?: string;
}

export interface AuthRenewManagerCredentialsInputDTO {
  email?: string;
  password?: string;
  provider_id?: number;
}

export interface AuthRenewManagerCredentialsRequest {
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  /** @default "user.login" */
  method: string;
  params?: AuthRenewManagerCredentialsInputDTO;
}

export interface AuthRenewManagerInputDTO {
  refresh_token?: string;
}

export interface AuthRenewManagerRefreshRequest {
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  /** @default "user.token_refresh" */
  method: string;
  params?: AuthRenewManagerInputDTO;
}

export interface AuthSSOAuthorizeInputDTO {
  client_id?: string;
  code_challenge?: string;
  code_challenge_method?: string;
  extra?: string;
  nonce?: string;
  path?: string;
  redirect_uri?: string;
  response_type?: string;
  scope?: string;
  state?: string;
  user_id?: number;
}

export interface AuthSSOAuthorizeOutputDTO {
  application?: string;
  client_id?: string;
  code?: string;
  extra?: string;
  nonce?: string;
  path?: string;
  redirect_uri?: string;
  scope?: string;
  state?: string;
}

export interface AuthSSOAuthorizeRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "sso.authorize" */
  method?: string;
  params?: AuthSSOAuthorizeInputDTO;
}

export interface AuthSSOAuthorizeResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: AuthSSOAuthorizeOutputDTO;
}

export interface AuthSSOError {
  error?: string;
  error_description?: string;
}

export interface AuthSSOJWKSOutputDTO {
  keys?: AuthJWK[];
}

export interface AuthSSOTokenIntrospect {
  active?: boolean;
  client_id?: string;
  exp?: number;
  iat?: number;
  scope?: string[];
  sub?: string;
  token_type?: string;
  username?: string;
}

export interface AuthSwaggerSSOToken {
  access_token?: string;
  expires_in?: number;
  id_token?: string;
  refresh_token?: string;
  state?: string;
  token_type?: string;
  user_id?: string;
}

export interface AuthSwaggerSSOTokenPublicData {
  /** Aud. Получатель токена (обычно client_id приложения, запрашивающего токен) */
  aud?: string;
  /** Azp. Конкретное приложение, которое инициировало запрос (обычно client_id приложения, запрашивающего токен) */
  azp?: string;
  /** Email. Почта уникальная пользователя */
  email?: string;
  /** Exp. Время истечения срока действия токена (в Unix timestamp) */
  exp?: number;
  /** Iat. Время выдачи токена (в Unix timestamp) */
  iat?: number;
  /** Iss. Идентификатор эмитента токена */
  iss?: string;
  /** K8S type */
  'k8s:access_type'?: string;
  /** LastLaunchId. ID пользователя SSO через LMS систему */
  last_launch_id?: string;
  /** Name. Полное ФИО пользователя */
  name?: string;
  /** Nonce.(Если запрос авторизации включал nonce) Случайное значение для предотвращения атак подмены */
  nonce?: string;
  /** Roles. Роли пользователя */
  roles?: string[];
  /** ServerID ID сервера аутентификации (как с Iss) */
  server_id?: number;
  /** Sub. Уникальный идентификатор пользователя в системе OpenID Provider (OP) */
  sub?: string;
  /** Username. Уникальный никнейм пользователя */
  username?: string;
}

export interface AuthSwaggerSSOTokenResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: AuthSwaggerSSOToken;
}

export interface AuthUserLogoutRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "user.logout" */
  method?: string;
  params?: any;
}

export interface AuthUserPasswordChangeInputDTO {
  again_password?: string;
  new_password?: string;
  old_password?: string;
}

export interface AuthUserPasswordChangeOutputDTO {
  id?: number;
}

export interface AuthUserPasswordChangeRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "user.password_change" */
  method?: string;
  params?: AuthUserPasswordChangeInputDTO;
}

export interface AuthUserPasswordChangeResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: AuthUserPasswordChangeOutputDTO;
}

export interface LtiQueryAuthProviderListInputDTO {
  is_auth?: boolean;
  limit?: number;
  offset?: number;
  search?: string;
}

export interface ModelsAuthProvider {
  base_uri: string;
  created_at: string;
  id?: number;
  key_set_uri?: string;
  lti_auth_login_uri?: string;
  lti_auth_token_uri?: string;
  lti_client_id?: string;
  lti_deployment_id?: string;
  name: string;
  private_key?: string;
  public_key?: string;
  sso_url?: string;
  target_link_uri?: string;
  type: string;
  updated_at: string;
}

export interface ModelsAuthProviderListItem {
  created_at: string;
  id?: number;
  name: string;
  sso_url?: string;
  type: string;
  updated_at: string;
}

export interface ModelsLTIAttempt {
  attempt_id: string;
  created_at: string;
  id?: number;
  lti_routing_id?: number;
  result?: object;
  room_id?: number;
  server_id?: number;
  status: string;
  synchronized_at?: string;
  updated_at: string;
  user_id?: number;
}

export interface ModelsLTIAttemptListItem {
  attempt_id: string;
  created_at: string;
  id?: number;
  lti_routing_id?: number;
  lti_routing_name?: string;
  repository?: string;
  result?: object;
  server_id?: number;
  server_name?: string;
  status: string;
  synchronized_at?: string;
  test_path?: string;
  updated_at: string;
  user_email?: string;
  user_id?: number;
  user_name?: string;
}

export interface ModelsLTIRouting {
  /** Параметры */
  collaboration?: number;
  created_at: string;
  id?: number;
  is_default?: boolean;
  /**
   * LabsPath holds the Git link of the laboratory with an optional `#ref`
   * fragment, for example https://git.example.org/course/lab.git#main.
   * The manifests are read from the repository root.
   */
  labs_path?: string;
  /**
   * The type of PNETLabsType, cms_client.PNETLabsTypeDefault
   * enum: default,curl,sso
   */
  labs_type?: string;
  lti_course_id?: string;
  lti_description?: string;
  lti_params_task?: string;
  lti_sub_id?: string;
  /** Автоматические */
  lti_task_id?: string;
  /** LTI Params */
  lti_title?: string;
  name?: string;
  server_id?: number;
  test_path?: string;
  updated_at: string;
}

export interface ModelsLTIRoutingListItem {
  created_at: string;
  id?: number;
  name?: string;
  updated_at: string;
}

export interface ModelsRole {
  code?: string;
  created_at: string;
  id?: number;
  name?: string;
  updated_at: string;
}

export interface ModelsServer {
  client_id?: string;
  created_at: string;
  id?: number;
  is_active?: boolean;
  last_count_users?: number;
  last_online_status?: string;
  max_count_users_limit?: number;
  minutes_for_disconnect?: number;
  name?: string;
  token?: string;
  /** enumeration: pnet, openid, k8s */
  type?: string;
  unit_rate?: number;
  updated_at: string;
  url?: string;
}

export interface ModelsServerListItem {
  created_at: string;
  id?: number;
  is_active?: boolean;
  last_count_users?: number;
  last_online_status?: string;
  max_count_users_limit?: number;
  minutes_for_disconnect?: number;
  name?: string;
  /** enumeration: pnet, openid, k8s */
  type?: string;
  unit_rate?: number;
  updated_at: string;
  url?: string;
}

export interface ModelsServiceCard {
  created_at: string;
  description: string;
  id?: number;
  image_url: string;
  is_active?: boolean;
  name: string;
  order?: number;
  updated_at: string;
  url: string;
}

export interface ModelsServiceCardListItem {
  created_at: string;
  description: string;
  id?: number;
  image_url: string;
  is_active?: boolean;
  name: string;
  order?: number;
  updated_at: string;
  url: string;
}

export interface ModelsTarget {
  created_at: string;
  description?: string;
  id?: string;
  /** список ссылок */
  internal_links?: object[];
  /** массив внутренних тегов */
  internal_tags?: string[];
  /** список ссылок */
  links?: object[];
  name?: string;
  synchronized_at: string;
  /** массив тегов */
  tags?: string[];
  type?: string;
  updated_at: string;
}

export interface ModelsTargetLink {
  type?: string;
  value?: string;
}

export interface ModelsTargetRelation {
  created_at: string;
  from_target_id?: string;
  id?: number;
  relation_type?: string;
  to_target_id?: string;
  updated_at: string;
}

export interface ModelsUser {
  created_at: string;
  deleted_at?: string;
  email?: string;
  group_name: string;
  id?: number;
  last_launch_id?: string;
  lti_user_id?: string;
  name?: string;
  store?: TypesJsonStore;
  updated_at: string;
}

export interface ModelsUserListItem {
  created_at: string;
  deleted_at?: string;
  email?: string;
  group_name: string;
  id?: number;
  lti_user_id?: string;
  name?: string;
  updated_at: string;
}

export interface QueriesLTIAttemptSearchParams {
  attempt_ids?: string[];
  limit?: number;
  offset?: number;
  server_client_ids?: string[];
  statuses?: string[];
  user_ids?: number[];
}

export interface QueriesRoundQueuePoolServerListItem {
  connected_at: string;
  id?: number;
  last_used: boolean;
  name?: string;
  type: string;
}

export interface QueriesServerQueriesListDTO {
  limit?: number;
  offset?: number;
  /**
   * The type of orderBy
   * enum: createdAt,unitRate,lastCountUsers
   */
  order_by?: string;
  search?: string;
  /**
   * The type of status
   * enum: all,active
   */
  status?: string;
  types?: string[];
}

export interface ServerQueueServerQueueListOutputDTO {
  model: QueriesRoundQueuePoolServerListItem[];
}

export interface ServerQueueServerQueueListRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "server_queue.list" */
  method?: string;
  params?: any;
}

export interface ServerQueueServerQueueListResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: ServerQueueServerQueueListOutputDTO;
}

export type ServerQueueServerQueueUpsertOutputDTO = object;

export interface ServerQueueServerQueueUpsertRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "server_queue.upsert" */
  method?: string;
  params?: any;
}

export interface ServerQueueServerQueueUpsertResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: ServerQueueServerQueueUpsertOutputDTO;
}

export type TypesJsonStore = Record<string, any>;

export interface TypesUserStoreGetRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "user.global_store_get" */
  method?: string;
  params?: any;
}

export interface TypesUserStoreGetResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: TypesJsonStore;
}

export interface TypesUserStoreSetRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "user.global_store_set" */
  method?: string;
  params?: TypesJsonStore;
}

export interface TypesUserStoreSetResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: boolean;
}

export interface UsecasesAuthProviderDeleteInputDTO {
  id?: number;
}

export interface UsecasesAuthProviderDeleteRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "lti_form.delete" */
  method?: string;
  params?: UsecasesLTIAttemptGetInputDTO;
}

export interface UsecasesAuthProviderDeleteResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesAuthProviderDeleteInputDTO;
}

export interface UsecasesAuthProviderEditInputDTO {
  auth_login_uri?: string;
  auth_token_uri?: string;
  base_uri: string;
  client_id?: string;
  deployment_id?: string;
  id?: number;
  key_set_uri?: string;
  name: string;
  sso_url?: string;
  target_link_uri?: string;
  type: string;
}

export interface UsecasesAuthProviderEditOutputDTO {
  id?: number;
}

export interface UsecasesAuthProviderEditRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "lti_form.upsert" */
  method?: string;
  params?: UsecasesAuthProviderEditInputDTO;
}

export interface UsecasesAuthProviderEditResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesAuthProviderEditOutputDTO;
}

export interface UsecasesAuthProviderGetInputDTO {
  id?: number;
}

export interface UsecasesAuthProviderGetOutputDTO {
  model?: ModelsAuthProvider;
}

export interface UsecasesAuthProviderGetRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "lti_form.get" */
  method?: string;
  params?: UsecasesAuthProviderGetInputDTO;
}

export interface UsecasesAuthProviderGetResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesAuthProviderGetOutputDTO;
}

export interface UsecasesAuthProviderListOutputDTO {
  model: ModelsAuthProviderListItem[];
  total_count: number;
}

export interface UsecasesAuthProviderListRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "lti_form.list" */
  method?: string;
  params?: LtiQueryAuthProviderListInputDTO;
}

export interface UsecasesAuthProviderListResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesAuthProviderListOutputDTO;
}

export interface UsecasesAvailableAddonInfo {
  id?: string;
  name?: string;
  type?: string;
}

export interface UsecasesConnectedAddonInfo {
  addon_id?: string;
  config?: Record<string, any>;
  id?: number;
  request_deleted_user_id?: number;
  type?: string;
}

export interface UsecasesLTIAttemptCreateInputDTO {
  room_number?: number;
}

export interface UsecasesLTIAttemptCreateOutputDTO {
  auto_redirect?: boolean;
  collaboration?: number;
  members?: ModelsUserListItem[];
  next_url?: string;
  room_number?: number;
}

export interface UsecasesLTIAttemptCreateRequest {
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  /** @default "lti_attempt.create" */
  method: string;
  params?: UsecasesLTIAttemptCreateInputDTO;
}

export interface UsecasesLTIAttemptCreateResponse {
  error?: any;
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  result?: UsecasesLTIAttemptCreateOutputDTO;
}

export interface UsecasesLTIAttemptDeleteInputDTO {
  id?: number;
}

export interface UsecasesLTIAttemptDeleteRequest {
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  /** @default "lti_attempt.delete" */
  method: string;
  params?: UsecasesLTIAttemptDeleteInputDTO;
}

export interface UsecasesLTIAttemptDeleteResponse {
  error?: any;
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  result?: UsecasesLTIAttemptDeleteInputDTO;
}

export interface UsecasesLTIAttemptEditBulkInput {
  attempt_id: string;
  result?: object;
  status?: string;
}

export interface UsecasesLTIAttemptEditBulkInputDTO {
  models?: UsecasesLTIAttemptEditBulkInput[];
}

export interface UsecasesLTIAttemptEditBulkOutputDTO {
  count?: number;
}

export interface UsecasesLTIAttemptEditBulkRequest {
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  /** @default "lti_attempt.update" */
  method: string;
  params?: UsecasesLTIAttemptEditBulkInputDTO;
}

export interface UsecasesLTIAttemptEditBulkResponse {
  error?: any;
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  result?: UsecasesLTIAttemptEditBulkOutputDTO;
}

export interface UsecasesLTIAttemptEditInputDTO {
  id?: number;
  result?: object;
  server_id?: number;
  status: string;
}

export interface UsecasesLTIAttemptEditOutputDTO {
  id?: number;
}

export interface UsecasesLTIAttemptEditRequest {
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  /** @default "lti_attempt.update" */
  method: string;
  params?: UsecasesLTIAttemptEditInputDTO;
}

export interface UsecasesLTIAttemptEditResponse {
  error?: any;
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  result?: UsecasesLTIAttemptEditOutputDTO;
}

export interface UsecasesLTIAttemptGetInputDTO {
  id?: number;
}

export interface UsecasesLTIAttemptGetOutputDTO {
  model?: ModelsLTIAttempt;
}

export interface UsecasesLTIAttemptGetRequest {
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  /** @default "lti_attempt.get" */
  method: string;
  params?: UsecasesLTIAttemptGetInputDTO;
}

export interface UsecasesLTIAttemptGetResponse {
  error?: any;
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  result?: UsecasesLTIAttemptGetOutputDTO;
}

export interface UsecasesLTIAttemptListOutputDTO {
  model: ModelsLTIAttemptListItem[];
  total_count: number;
}

export interface UsecasesLTIAttemptListRequest {
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  /** @default "lti_attempt.list" */
  method: string;
  params?: QueriesLTIAttemptSearchParams;
}

export interface UsecasesLTIAttemptListResponse {
  error?: any;
  /** @default "1" */
  id: string;
  /** @default "2.0" */
  jsonrpc: string;
  result?: UsecasesLTIAttemptListOutputDTO;
}

export interface UsecasesLTIRoutingDeleteInputDTO {
  id?: number;
}

export interface UsecasesLTIRoutingDeleteRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "lti_routing.delete" */
  method?: string;
  params?: UsecasesLTIRoutingDeleteInputDTO;
}

export interface UsecasesLTIRoutingDeleteResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesLTIRoutingDeleteInputDTO;
}

export interface UsecasesLTIRoutingEditInputDTO {
  collaboration?: number;
  id?: number;
  is_default?: boolean;
  labs_path?: string;
  labs_type?: string;
  lti_course_id?: string;
  lti_description?: string;
  lti_params_task?: string;
  lti_sub_id?: string;
  lti_task_id?: string;
  lti_title?: string;
  name?: string;
  server_id?: number;
  test_path?: string;
}

export interface UsecasesLTIRoutingEditOutputDTO {
  id?: number;
}

export interface UsecasesLTIRoutingEditRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "lti_routing.upsert" */
  method?: string;
  params?: UsecasesLTIRoutingEditInputDTO;
}

export interface UsecasesLTIRoutingEditResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesLTIRoutingEditOutputDTO;
}

export interface UsecasesLTIRoutingGetInputDTO {
  id?: number;
}

export interface UsecasesLTIRoutingGetOutputDTO {
  model?: ModelsLTIRouting;
}

export interface UsecasesLTIRoutingGetRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "lti_routing.get" */
  method?: string;
  params?: UsecasesLTIRoutingGetInputDTO;
}

export interface UsecasesLTIRoutingGetResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesLTIRoutingGetOutputDTO;
}

export interface UsecasesLTIRoutingListInputDTO {
  limit?: number;
  offset?: number;
  search?: string;
}

export interface UsecasesLTIRoutingListOutputDTO {
  model: ModelsLTIRoutingListItem[];
  total_count: number;
}

export interface UsecasesLTIRoutingListRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "lti_routing.list" */
  method?: string;
  params?: UsecasesLTIRoutingListInputDTO;
}

export interface UsecasesLTIRoutingListResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesLTIRoutingListOutputDTO;
}

export interface UsecasesRoleDeleteInputDTO {
  id: number;
}

export interface UsecasesRoleDeleteRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "role.delete" */
  method?: string;
  params?: UsecasesRoleDeleteInputDTO;
}

export interface UsecasesRoleDeleteResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesRoleDeleteInputDTO;
}

export interface UsecasesRoleEditInputDTO {
  code?: string;
  id?: number;
  name?: string;
}

export interface UsecasesRoleEditOutputDTO {
  id?: number;
}

export interface UsecasesRoleEditRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "role.upsert" */
  method?: string;
  params?: UsecasesRoleEditInputDTO;
}

export interface UsecasesRoleEditResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesRoleEditOutputDTO;
}

export type UsecasesRoleListInputDTO = object;

export interface UsecasesRoleListOutputDTO {
  model: ModelsRole[];
  total_count: number;
}

export interface UsecasesRoleListRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "role.list" */
  method?: string;
  params?: UsecasesRoleListInputDTO;
}

export interface UsecasesRoleListResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesRoleListOutputDTO;
}

export interface UsecasesServerDeleteInputDTO {
  id?: number;
}

export interface UsecasesServerDeleteRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "server.list" */
  method?: string;
  params?: UsecasesServerDeleteInputDTO;
}

export interface UsecasesServerDeleteResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesServerDeleteInputDTO;
}

export interface UsecasesServerEditInputDTO {
  client_id?: string;
  id?: number;
  is_active?: boolean;
  max_count_users_limit?: number;
  minutes_for_disconnect?: number;
  name: string;
  roles?: number[];
  token?: string;
  type: string;
  unit_rate?: number;
  url: string;
}

export interface UsecasesServerEditOutputDTO {
  id?: number;
}

export interface UsecasesServerEditRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "server.upsert" */
  method?: string;
  params?: UsecasesServerEditInputDTO;
}

export interface UsecasesServerEditResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesServerEditOutputDTO;
}

export interface UsecasesServerGetInputDTO {
  id?: number;
}

export interface UsecasesServerGetOutputDTO {
  model?: ModelsServer;
  roles?: number[];
}

export interface UsecasesServerGetRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "server.get" */
  method?: string;
  params?: UsecasesServerGetInputDTO;
}

export interface UsecasesServerGetResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesServerGetOutputDTO;
}

export interface UsecasesServerListModel {
  model: ModelsServerListItem;
  roles?: number[];
}

export interface UsecasesServerListOutputDTO {
  model: UsecasesServerListModel[];
  total_count: number;
}

export interface UsecasesServerListRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "server.list" */
  method?: string;
  params?: QueriesServerQueriesListDTO;
}

export interface UsecasesServerListResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesServerListOutputDTO;
}

export interface UsecasesServiceCardDeleteInputDTO {
  id?: number;
}

export interface UsecasesServiceCardDeleteRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.list" */
  method?: string;
  params?: UsecasesServiceCardDeleteInputDTO;
}

export interface UsecasesServiceCardDeleteResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesServiceCardDeleteInputDTO;
}

export interface UsecasesServiceCardEditInputDTO {
  description?: string;
  id?: number;
  image_url?: string;
  is_active?: boolean;
  name?: string;
  order?: number;
  url?: string;
}

export interface UsecasesServiceCardEditRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.upsert" */
  method?: string;
  params?: UsecasesServiceCardEditInputDTO;
}

export interface UsecasesServiceCardEditResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesServiceCardDeleteInputDTO;
}

export interface UsecasesServiceCardGetInputDTO {
  id?: number;
}

export interface UsecasesServiceCardGetOutputDTO {
  model?: ModelsServiceCard;
}

export interface UsecasesServiceCardGetRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.get" */
  method?: string;
  params?: UsecasesServiceCardGetInputDTO;
}

export interface UsecasesServiceCardGetResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesServiceCardGetOutputDTO;
}

export interface UsecasesServiceCardListInputDTO {
  limit?: number;
  offset?: number;
  search?: string;
}

export interface UsecasesServiceCardListOutputDTO {
  model: ModelsServiceCardListItem[];
  total_count: number;
}

export interface UsecasesServiceCardListRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.list" */
  method?: string;
  params?: UsecasesServiceCardListInputDTO;
}

export interface UsecasesServiceCardListResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesServiceCardListOutputDTO;
}

export interface UsecasesTargetAddonCreateInputDTO {
  addon_id: string;
  iss_id?: string;
  name?: string;
  target_id: string;
}

export interface UsecasesTargetAddonCreateOutputDTO {
  id?: number;
}

export interface UsecasesTargetAddonCreateRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.upsert" */
  method?: string;
  params?: UsecasesTargetAddonCreateInputDTO;
}

export interface UsecasesTargetAddonCreateResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesTargetAddonCreateOutputDTO;
}

export interface UsecasesTargetAddonDeleteInputDTO {
  addon_id: string;
  iss_id?: string;
  revoke?: boolean;
  target_id: string;
}

export interface UsecasesTargetAddonDeleteOutputDTO {
  message?: string;
  pending_confirmation?: boolean;
}

export interface UsecasesTargetAddonResetInputDTO {
  addon_id: string;
  iss_id?: string;
  target_id: string;
}

export interface UsecasesTargetAddonResetOutputDTO {
  id?: number;
}

export interface UsecasesTargetDeleteInputDTO {
  id: string;
}

export type UsecasesTargetDeleteOutputDTO = object;

export interface UsecasesTargetGetInputDTO {
  id: string;
}

export interface UsecasesTargetGetOutputDTO {
  availableAddons?: UsecasesAvailableAddonInfo[];
  connectedAddons?: UsecasesConnectedAddonInfo[];
  created_at: string;
  description?: string;
  id?: string;
  /** список ссылок */
  internal_links?: object[];
  /** массив внутренних тегов */
  internal_tags?: string[];
  /** список ссылок */
  links?: object[];
  name?: string;
  synchronized_at: string;
  /** массив тегов */
  tags?: string[];
  targetUsers?: UsecasesTargetUserInfo[];
  type?: string;
  updated_at: string;
}

export interface UsecasesTargetGetRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.upsert" */
  method?: string;
  params?: UsecasesTargetGetInputDTO;
}

export interface UsecasesTargetGetResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesTargetGetOutputDTO;
}

export interface UsecasesTargetItem {
  is_mine: boolean;
  taget: ModelsTarget;
}

export type UsecasesTargetListInputDTO = object;

export interface UsecasesTargetListOutputDTO {
  model?: UsecasesTargetModel;
}

export interface UsecasesTargetListRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.upsert" */
  method?: string;
  params?: UsecasesTargetListInputDTO;
}

export interface UsecasesTargetListResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesTargetListOutputDTO;
}

export interface UsecasesTargetModel {
  relations?: UsecasesTargetRelationItem[];
  targets?: UsecasesTargetItem[];
}

export interface UsecasesTargetRelationCreateInputDTO {
  from_target_id: string;
  relation_type: string;
  to_target_id: string;
}

export interface UsecasesTargetRelationCreateOutputDTO {
  id?: number;
}

export interface UsecasesTargetRelationCreateRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.upsert" */
  method?: string;
  params?: UsecasesTargetRelationCreateInputDTO;
}

export interface UsecasesTargetRelationCreateResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesTargetRelationCreateOutputDTO;
}

export interface UsecasesTargetRelationDeleteInputDTO {
  from_target_id?: string;
  relation_type?: string;
  to_target_id?: string;
}

export interface UsecasesTargetRelationDeleteRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.upsert" */
  method?: string;
  params?: UsecasesTargetRelationDeleteInputDTO;
}

export interface UsecasesTargetRelationDeleteResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesTargetRelationDeleteInputDTO;
}

export interface UsecasesTargetRelationItem {
  relation?: ModelsTargetRelation;
}

export interface UsecasesTargetUpsertInputDTO {
  description?: string;
  /** nil – создание, иначе – обновление */
  id?: string;
  links?: ModelsTargetLink[];
  name: string;
  tags?: string[];
  type: string;
}

export interface UsecasesTargetUpsertOutputDTO {
  id?: string;
}

export interface UsecasesTargetUpsertRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.upsert" */
  method?: string;
  params?: UsecasesTargetUpsertInputDTO;
}

export interface UsecasesTargetUpsertResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesTargetUpsertOutputDTO;
}

export interface UsecasesTargetUserDeleteInputDTO {
  target_id: string;
  user_id: number;
}

export type UsecasesTargetUserDeleteOutputDTO = object;

export interface UsecasesTargetUserDeleteRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.upsert" */
  method?: string;
  params?: UsecasesTargetUserDeleteInputDTO;
}

export interface UsecasesTargetUserDeleteResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesTargetUserDeleteOutputDTO;
}

export interface UsecasesTargetUserInfo {
  roles?: string[];
  user_id?: number;
}

export interface UsecasesTargetUserUpsertInputDTO {
  roles: string[];
  target_id: string;
  user_id: number;
}

export interface UsecasesTargetUserUpsertOutputDTO {
  id?: number;
}

export interface UsecasesTargetUserUpsertRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "service_card.upsert" */
  method?: string;
  params?: UsecasesTargetUserUpsertInputDTO;
}

export interface UsecasesTargetUserUpsertResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesTargetUserUpsertOutputDTO;
}

export interface UsecasesUserEditInputDTO {
  email: string;
  group_name?: string;
  id?: number;
  is_active?: boolean;
  lti_user_id?: string;
  name: string;
  roles?: number[];
  store?: TypesJsonStore;
}

export interface UsecasesUserEditOutputDTO {
  id?: number;
}

export interface UsecasesUserEditRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "user.upsert" */
  method?: string;
  params?: UsecasesUserEditInputDTO;
}

export interface UsecasesUserEditResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesUserEditOutputDTO;
}

export interface UsecasesUserGetInputDTO {
  id?: number;
}

export interface UsecasesUserGetOutputDTO {
  model?: ModelsUser;
  roles?: number[];
}

export interface UsecasesUserGetRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "user.get" */
  method?: string;
  params?: UsecasesUserGetInputDTO;
}

export interface UsecasesUserGetResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesUserGetOutputDTO;
}

export interface UsecasesUserListInputDTO {
  limit?: number;
  offset?: number;
  search?: string;
  user_ids?: number[];
}

export interface UsecasesUserListModel {
  model: ModelsUserListItem;
  roles?: number[];
}

export interface UsecasesUserListOutputDTO {
  model: UsecasesUserListModel[];
  total_count: number;
}

export interface UsecasesUserListRequest {
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  /** @default "user.get" */
  method?: string;
  params?: UsecasesUserListInputDTO;
}

export interface UsecasesUserListResponse {
  error?: any;
  /** @default "1" */
  id?: string;
  /** @default "2.0" */
  jsonrpc?: string;
  result?: UsecasesUserListOutputDTO;
}
