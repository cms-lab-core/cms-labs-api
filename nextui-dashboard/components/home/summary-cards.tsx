import { Activity, CircleStop, Clock3, FlaskConical, ShieldCheck } from 'lucide-react';
import { Link } from 'react-router-dom';
import { ContentCardWrapperMain } from './card-wrapper';
import { Loading } from '@/components/scroll/loader';
import { RoutesLocation } from '@/components/routes';
import useLanguageBrowser from '@/helpers/locale';
import { useQueryLtiAttemptList } from '@/helpers/queries/lti_attempt/use-query-lti-attempt-list';
import { parseAttemptCheckResult } from '@/components/pages/lti-attempts/table/check-result';

const activeStatuses = ['pending', 'active', 'terminating'];

export const HomeSummaryCards = () => {
  const {
    locale: { Home }
  } = useLanguageBrowser();
  const response = useQueryLtiAttemptList({ limit: 500, statuses: activeStatuses }, { refetchInterval: 60_000 });
  const attempts = response.data?.model || [];
  const running = attempts.filter((attempt) => attempt.status === 'active').length;
  const pending = attempts.filter((attempt) => attempt.status === 'pending').length;
  const terminating = attempts.filter((attempt) => attempt.status === 'terminating').length;
  const checkResults = attempts.map((attempt) => parseAttemptCheckResult(attempt.result)).filter(Boolean);
  const checked = checkResults.length;
  const passedTasks = checkResults.reduce((sum, result) => sum + (result?.passedTasks || 0), 0);
  const totalTasks = checkResults.reduce((sum, result) => sum + (result?.tasks.length || 0), 0);

  const cards = [
    {
      label: Home.SessionsTotal,
      value: attempts.length,
      description: Home.SessionsTotalDescription,
      icon: Activity,
      className: 'text-primary',
      href: RoutesLocation.ltiAttempts({ statuses: activeStatuses })
    },
    {
      label: Home.SessionsRunning,
      value: running,
      description: Home.SessionsRunningDescription,
      icon: FlaskConical,
      className: 'text-success',
      href: RoutesLocation.ltiAttempts({ statuses: ['active'] })
    },
    {
      label: Home.SessionsPending,
      value: pending,
      description: Home.SessionsPendingDescription,
      icon: Clock3,
      className: 'text-warning',
      href: RoutesLocation.ltiAttempts({ statuses: ['pending'] })
    },
    {
      label: Home.SessionsChecked,
      value: checked,
      description:
        totalTasks > 0 ? `${Home.ChecksPassed}: ${passedTasks} / ${totalTasks}` : Home.SessionsCheckedDescription,
      icon: ShieldCheck,
      className: 'text-secondary',
      href: RoutesLocation.ltiAttempts({ statuses: activeStatuses })
    },
    {
      label: Home.SessionsTerminating,
      value: terminating,
      description: Home.SessionsTerminatingDescription,
      icon: CircleStop,
      className: 'text-danger',
      href: RoutesLocation.ltiAttempts({ statuses: ['terminating'] })
    }
  ];

  return (
    <ContentCardWrapperMain
      title={Home.OverviewTitle}
      link={RoutesLocation.ltiAttempts({ statuses: activeStatuses })}
      wrapChildren
    >
      {response.isLoading ? (
        <Loading size='md' />
      ) : (
        <div>
          <p className='mb-5 text-sm text-default-500'>{Home.OverviewDescription}</p>
          <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-5'>
            {cards.map((card) => {
              const Icon = card.icon;
              return (
                <Link
                  key={card.label}
                  to={card.href}
                  className='group rounded-xl border border-divider bg-content1 p-4 transition-colors hover:bg-default-100'
                >
                  <div className='flex items-start justify-between gap-3'>
                    <div className='min-w-0'>
                      <p className='text-xs font-medium text-default-500'>{card.label}</p>
                      <p className='mt-2 font-mono text-2xl font-semibold text-foreground'>{card.value}</p>
                    </div>
                    <Icon className={`h-5 w-5 shrink-0 ${card.className}`} />
                  </div>
                  <p className='mt-3 text-xs leading-5 text-default-400'>{card.description}</p>
                </Link>
              );
            })}
          </div>
          {response.dataUpdatedAt > 0 && (
            <p className='mt-4 text-right text-xs text-default-400'>
              {Home.UpdatedAt}: {new Date(response.dataUpdatedAt).toLocaleTimeString()}
            </p>
          )}
        </div>
      )}
    </ContentCardWrapperMain>
  );
};
