import { Divider } from '@heroui/divider';
import useLanguageBrowser from '@/helpers/locale';
import { Link } from 'react-router-dom';
import { RoutesLocation } from '@/components/routes';

const defaultCredentials = {
  email: 'admin@admin.com',
  password: 'admin'
} as const;

interface Props {
  children: React.ReactNode;
}

export const AuthLayoutWrapper = ({ children }: Props) => {
  const { locale } = useLanguageBrowser();
  return (
    <div className='flex h-screen'>
      <div className='flex-1 flex-col flex items-center justify-center p-6'>{children}</div>

      <div className='hidden my-10 md:block'>
        <Divider orientation='vertical' />
      </div>

      <div className='hidden md:flex flex-1 relative flex items-center justify-center p-6'>
        <div className='z-10'>
          <h1 className='font-bold text-[45px]'>{locale.Auth.MainTitle}</h1>
          <div className='font-light text-slate-400 mt-4 mb-4'>{locale.Auth.MainDescription}</div>
          <div
            role='note'
            className='w-full max-w-md mb-6 rounded-xl border border-primary-200 bg-primary-50 p-4 text-sm'
          >
            <div className='font-semibold text-primary-700'>{locale.Login.InitialCredentialsTitle}</div>
            <div className='mt-1 text-default-600'>{locale.Login.InitialCredentialsDescription}</div>
            <dl className='mt-3 grid grid-cols-[auto_1fr] gap-x-3 gap-y-1'>
              <dt className='text-default-500'>{locale.Login.InitialCredentialsLogin}:</dt>
              <dd>
                <code>{defaultCredentials.email}</code>
              </dd>
              <dt className='text-default-500'>{locale.Login.InitialCredentialsPassword}:</dt>
              <dd>
                <code>{defaultCredentials.password}</code>
              </dd>
            </dl>
          </div>
          <div className='font-light underline text-slate-400'>
            <Link to={RoutesLocation.language()}>{locale.Auth.MainChangeLanguage}</Link>
          </div>
        </div>
      </div>
    </div>
  );
};
