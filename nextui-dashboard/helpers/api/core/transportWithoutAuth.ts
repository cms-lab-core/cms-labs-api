import { RpcTransport } from './request';
import { AuthSwaggerSSOToken } from '@/helpers/api/types';
import { AuthData } from '@/helpers/api/core/types';
import { CamelCasedPropertiesDeep } from 'type-fest';

export const CoreJsonRpcPath = '/api/v1/rpc';
export const ClabgateJsonRpcPath = '/clabgate/api/v1/rpc';

class TransportWithoutAuth {
  public t: RpcTransport;
  private authData?: AuthData;

  constructor() {
    this.t = new RpcTransport();
    this.authData = {
      accessToken: '',
      expiredAt: new Date(-1)
    };
  }

  async userTokenRefresh(): Promise<void> {
    const response = await this.t.rpc<CamelCasedPropertiesDeep<AuthSwaggerSSOToken>>(CoreJsonRpcPath, {
      method: 'user.token_refresh',
      params: {}
    });
    this.authData = {
      accessToken: response.accessToken,
      expiredAt: new Date((response.expiresIn ?? 0) * 1000)
    };
  }

  userTokenAccess(): AuthData | undefined {
    return this.authData;
  }

  userTokenClear(): void {
    this.authData = {
      accessToken: '',
      expiredAt: new Date(-1)
    };
  }
}

export const transportWithoutAuth = new TransportWithoutAuth();
