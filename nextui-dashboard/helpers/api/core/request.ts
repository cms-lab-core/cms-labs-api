import { nanoid } from 'nanoid';
import { objectToCamel, objectToSnake } from 'ts-case-convert';
import axios, { AxiosError, type AxiosRequestConfig, type AxiosResponse } from 'axios';
import type { AuthDataGet, AuthDataUpdate, AuthOptionsSetup, RpcParams } from './types';

function onceTime(func: () => Promise<void>) {
  let promise: Promise<void> | null = null;

  return async (): Promise<void> => {
    if (!promise)
      promise = func().finally(() => {
        promise = null;
      });
    await promise;
  };
}

function tokenIsExpired(expiredAt: Date): boolean {
  return Date.now() > expiredAt.getTime() - 60 * 1000;
}

export class RpcTransport {
  private getAuthData: AuthDataGet | undefined;
  private updateAuthData: AuthDataUpdate | undefined;
  private transport;

  constructor() {
    this.transport = axios.create({
      baseURL: window.location.origin,
      headers: { 'Content-Type': 'application/json' },
      withCredentials: true
    });

    this.setupInterceptors();
  }

  setupAuth(options: AuthOptionsSetup): void {
    this.getAuthData = options.getAuthData;
    this.updateAuthData = onceTime(options.updateAuthData);
  }

  private setupInterceptors(): void {
    this.transport.interceptors.request.use(
      // @ts-ignore
      async (config: AxiosRequestConfig) => {
        const expiredAt = this.getAuthData && this.getAuthData()?.expiredAt;

        if (expiredAt && this.updateAuthData && tokenIsExpired(expiredAt)) {
          await this.updateAuthData();
        }
        const token = this.getAuthData?.()?.accessToken;
        let xRequestId = nanoid();
        try {
          xRequestId = config.data?.id;
        } catch (e) {
          console.debug('x-request-id not parse from body.id jsonrpc protocol');
        }
        config.headers = {
          ...config.headers,
          ...(!!token && { Authorization: `Bearer ${token}` }),
          'x-request-id': xRequestId
        };
        return config;
      }
    );
    this.transport.interceptors.response.use(
      (response: AxiosResponse) => {
        if (response.data?.error) throw response.data.error;
        return response;
      },
      (error: AxiosError) => {
        // @ts-ignore
        if (error?.code === 'ERR_BAD_RESPONSE' && error.response?.data?.error) {
          // @ts-ignore
          throw error.response?.data?.error;
        }
        return Promise.reject(error);
      }
    );
  }

  /**
   * JsonRpc (post) для общения с сервером (исключая запрос авторизации).
   * Автоматически преобразует данные запроса и ответа.
   */
  async rpc<Response>(url: string, data: RpcParams | object, config?: AxiosRequestConfig): Promise<Response> {
    const rpcData = { ...data, id: nanoid(), jsonrpc: '2.0' };
    const params = objectToSnake(rpcData);

    return this.transport.post(url, params, config).then((resp) => {
      return objectToCamel(resp.data.result) as Response;
    });
  }

  /**
   * Получение внутреннего транспорта для прямого доступа
   * (если нужны специфичные методы axios)
   */
  getTransport() {
    return this.transport;
  }

  static getQueryKey(url: string, method: string, params: object): any {
    return [url, method, params];
  }

  getQueryOptions<Response, Params extends object, Options extends object = object>(
    url: string,
    method: string,
    params: Params,
    options?: Options,
    config?: AxiosRequestConfig
  ) {
    return {
      queryKey: RpcTransport.getQueryKey(url, method, params),
      queryFn: ({ signal }: { signal?: AbortSignal }) =>
        this.rpc<Response>(url, { method, params }, { signal, ...config }),
      ...options
    };
  }

  /** Получение ключа на основе url, method и params с добавлением infinite */
  static getInfiniteQueryKey(url: string, method: string, params: object): any {
    return [url, method, 'infinite', params];
  }
}
