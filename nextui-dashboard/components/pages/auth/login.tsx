'use client';

import { LoginSchema } from '@/helpers/schemas';
import { Button, Input, Tab, Tabs } from '@heroui/react';
import { Formik } from 'formik';
import useLanguageBrowser from '@/helpers/locale';
import { RoutesLocation } from '@/components/routes';
import { SSOAuthorizationGet, SSOAuthorizationReset } from '@/components/pages/auth/ssoSave';
import { SSOAuthorizationComplete } from '@/components/pages/auth/sso';
import { useMutationUserLogin } from '@/helpers/queries/user/use-mutation-user-login';
import { useMutationSsoAuthorize } from '@/helpers/queries/sso/use-mutation-sso-authorize';
import { CamelCasedPropertiesDeep } from 'type-fest';
import { AuthRenewManagerCredentialsInputDTO } from '@/helpers/api';
import { useSearchParams } from 'react-router-dom';
import { useMemo } from 'react';
import { KeyRound } from 'lucide-react';
import { useQueryAuthProviderList } from '@/helpers/queries/lti_form/use-query-lti-form-list';

const safeReturnTo = (returnTo: string | null): string | null => {
  if (!returnTo?.startsWith('/') || returnTo.startsWith('//')) return null;
  return returnTo;
};

const defaultValues: CamelCasedPropertiesDeep<AuthRenewManagerCredentialsInputDTO> = {
  email: '',
  password: '',
  providerId: 0
};

export const Login = () => {
  const { locale } = useLanguageBrowser();
  const [searchParams, setSearchParams] = useSearchParams();
  const returnTo = safeReturnTo(searchParams.get('return_to'));

  const { data: authProvidersData, isLoading } = useQueryAuthProviderList({
    isAuth: true,
    limit: 100,
    offset: 0
  });

  const ltiProviders = authProvidersData?.model?.filter((p) => !!p.ssoUrl) || [];
  const ldapProviders = authProvidersData?.model?.filter((p) => p.type === 'ldap') || [];

  const tabs = useMemo(() => {
    const tabs = [];
    if (ltiProviders.length > 0) {
      tabs.push({ id: 'sso', label: locale.Login.TabSSO, providerId: 0 });
    }
    ldapProviders.forEach((p) => {
      tabs.push({ id: `ldap:${p.id}`, label: p.name || 'LDAP', providerId: p.id });
    });
    tabs.push({ id: 'internal', label: locale.Login.TabInternal, providerId: 0 });
    return tabs;
  }, [authProvidersData, isLoading]);

  const activeTab = searchParams.get('type') || tabs?.[0]?.id;

  const { mutate: mutateSsoAuthorize } = useMutationSsoAuthorize({
    onSuccess: (data) => {
      SSOAuthorizationReset();
      const redirectUrl = SSOAuthorizationComplete(data);
      window.location.href = redirectUrl;
    }
  });

  const { mutate } = useMutationUserLogin({
    onSuccess: (data, { formikHelpers }) => {
      formikHelpers.resetForm();
      const params = SSOAuthorizationGet();
      if (params === null) {
        window.location.href = returnTo || RoutesLocation.home();
      } else {
        mutateSsoAuthorize({
          clientId: params.clientId,
          redirectUri: params.redirectUri,
          responseType: params.responseType,
          scope: params.scope,
          path: params.path,
          state: params.state,
          nonce: params.nonce,
          extra: params.extra,
          codeChallenge: params.codeChallenge,
          codeChallengeMethod: params.codeChallengeMethod
        });
      }
    },
    onError: (error: any, { formikHelpers }) => {
      formikHelpers.setErrors({});
      formikHelpers.setErrors({ password: error?.data?.message });
    }
  });

  return (
    <div className='flex flex-col items-center w-full'>
      <div className='text-center text-[25px] font-bold mb-6'>{locale.Login.PageName}</div>
      <Tabs
        selectedKey={activeTab}
        onSelectionChange={(key) => {
          const nextSearchParams = new URLSearchParams(searchParams);
          nextSearchParams.set('type', key as string);
          setSearchParams(nextSearchParams);
        }}
        className='mb-6'
      >
        {tabs.map((tab, index) => (
          <Tab key={tab.id} title={tab.label} />
        ))}
      </Tabs>

      <div className='w-full max-w-md'>
        <Formik
          initialValues={defaultValues}
          validationSchema={LoginSchema()}
          onSubmit={(values, formikHelpers) => {
            const provider = tabs.find((s) => s.id === activeTab);
            // Вызываем мутацию только для internal/ldap табов
            if (activeTab !== 'sso') {
              mutate({
                values: {
                  email: values.email,
                  password: values.password,
                  providerId: provider?.providerId || 0
                },
                // @ts-ignore
                formikHelpers
              });
            }
          }}
        >
          {({ values, errors, touched, handleChange, handleSubmit }) => (
            <div>
              {activeTab === 'sso' ? (
                // SSO кнопки
                <div className='flex flex-col gap-3'>
                  {ltiProviders.map((service) => (
                    <Button
                      key={service.id}
                      startContent={<KeyRound className='h-5 w-5 stroke-[#969696]' />}
                      className='w-full'
                      onPress={() => {
                        window.location.href = service.ssoUrl || '';
                      }}
                      variant='flat'
                      color='default'
                    >
                      {service.name}
                    </Button>
                  ))}
                </div>
              ) : (
                // Форма логина (internal/ldap)
                <>
                  <Input
                    className='w-full'
                    variant='bordered'
                    label={locale.Login.FieldEmail}
                    description={locale.Login.FieldEmailDescription}
                    type='email'
                    value={values.email}
                    isInvalid={!!errors.email && !!touched.email}
                    errorMessage={errors.email}
                    onChange={handleChange('email')}
                  />
                  <Input
                    className='w-full'
                    variant='bordered'
                    label={locale.Login.FieldPassword}
                    type='password'
                    value={values.password}
                    isInvalid={!!errors.password && !!touched.password}
                    errorMessage={errors.password}
                    onChange={handleChange('password')}
                  />
                  <Button className='w-full mt-3' onPress={() => handleSubmit()} variant='flat' color='default'>
                    {locale.Login.Submit}
                  </Button>
                </>
              )}
            </div>
          )}
        </Formik>
      </div>
    </div>
  );
};
