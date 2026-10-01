import { useEffect, useState } from 'react';
import {
  Clock3,
  Eraser,
  GraduationCap,
  LayoutDashboard,
  LoaderCircle,
  Menu,
  Moon,
  Network,
  NotebookTabs,
  RefreshCw,
  RotateCcw,
  Square,
  Sun,
  UserRound
} from 'lucide-react';
import { Button, Dropdown, DropdownItem, DropdownMenu, DropdownTrigger } from '@heroui/react';
import type { LabCopy } from './copy';
import { formatDuration, phaseTone, type LabSession, type WorkspaceTab } from './types';
import type { LanguageType } from '@/helpers/locale/locale';
import type { ThemeType } from '@/components/navbar/useTheme';

interface WorkspaceHeaderProps {
  copy: LabCopy;
  session?: LabSession;
  activeTab: WorkspaceTab;
  userName?: string;
  refreshing: boolean;
  stopping: boolean;
  topologyActionPending: boolean;
  topologyActionsAvailable: boolean;
  language: LanguageType;
  theme: ThemeType;
  onTabChange: (tab: WorkspaceTab) => void;
  onRefresh: () => void;
  onStop: () => void;
  onHome: () => void;
  onToggleLanguage: () => void;
  onToggleTheme: () => void;
  onRestartTopology: () => void;
  onWipeTopology: () => void;
  onStartTour: () => void;
}

const tabs = [
  { id: 'overview' as const, icon: LayoutDashboard, key: 'overview' as const },
  { id: 'topology' as const, icon: Network, key: 'topology' as const },
  { id: 'assignment' as const, icon: NotebookTabs, key: 'assignment' as const }
];

export const WorkspaceHeader = ({
  copy,
  session,
  activeTab,
  userName,
  refreshing,
  stopping,
  topologyActionPending,
  topologyActionsAvailable,
  language,
  theme,
  onTabChange,
  onRefresh,
  onStop,
  onHome,
  onToggleLanguage,
  onToggleTheme,
  onRestartTopology,
  onWipeTopology,
  onStartTour
}: WorkspaceHeaderProps) => {
  const [, setClock] = useState(0);
  useEffect(() => {
    const timer = window.setInterval(() => setClock((value) => value + 1), 1000);
    return () => window.clearInterval(timer);
  }, []);

  const tone = phaseTone(session?.phase);
  const status =
    tone === 'success'
      ? copy.running
      : tone === 'danger'
        ? copy.failed
        : tone === 'warning'
          ? session?.phase === 'stopping'
            ? copy.stopping
            : copy.degraded
          : copy.preparing;
  const statusColor =
    tone === 'success'
      ? 'bg-emerald-400'
      : tone === 'danger'
        ? 'bg-red-400'
        : tone === 'warning'
          ? 'bg-amber-400'
          : 'bg-blue-400';
  const initials = (userName || session?.username || 'U')
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join('');

  return (
    <header className='flex h-14 shrink-0 items-center border-b border-slate-700/80 bg-[#151f30] text-slate-100'>
      <nav
        data-tour='tabs'
        className='flex h-full min-w-0 flex-1 items-end overflow-x-auto px-2'
        aria-label={copy.labStand}
      >
        {tabs.map((tab) => {
          const Icon = tab.icon;
          const active = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              data-workspace-tab={tab.id}
              type='button'
              onClick={() => onTabChange(tab.id)}
              className={`relative flex h-full min-w-max items-center gap-2 px-3 text-xs font-medium transition sm:px-4 ${
                active ? 'bg-slate-700/30 text-slate-50' : 'text-slate-500 hover:bg-slate-700/20 hover:text-slate-300'
              }`}
            >
              <Icon className={`h-4 w-4 ${active ? 'text-emerald-400' : ''}`} />
              {copy[tab.key]}
              {active && <span className='absolute inset-x-2 bottom-0 h-0.5 bg-emerald-400' />}
            </button>
          );
        })}
      </nav>

      <div className='flex h-full shrink-0 items-center gap-1 border-l border-slate-700/80 px-2 sm:gap-2 sm:px-3'>
        <div className='hidden items-center gap-2 rounded-md border border-slate-700 bg-slate-800/50 px-2.5 py-1.5 text-xs lg:flex'>
          <span className={`h-1.5 w-1.5 rounded-full ${statusColor} ${tone === 'info' ? 'animate-pulse' : ''}`} />
          <span className='text-slate-300'>{status}</span>
        </div>
        <div className='hidden items-center gap-1.5 font-mono text-xs text-slate-500 xl:flex'>
          <Clock3 className='h-3.5 w-3.5' />
          {formatDuration(session?.createdAt)}
        </div>
        <Button
          isIconOnly
          size='sm'
          radius='sm'
          variant='light'
          onPress={onRefresh}
          title={copy.refresh}
          className='text-slate-500'
        >
          <RefreshCw className={`h-4 w-4 ${refreshing ? 'animate-spin' : ''}`} />
        </Button>
        <Button
          isIconOnly
          size='sm'
          radius='sm'
          variant='light'
          onPress={onToggleTheme}
          title={theme === 'dark' ? copy.switchToLight : copy.switchToDark}
          className='text-slate-500'
        >
          {theme === 'dark' ? <Sun className='h-4 w-4' /> : <Moon className='h-4 w-4' />}
        </Button>
        <Button
          size='sm'
          radius='sm'
          variant='light'
          onPress={onToggleLanguage}
          title={copy.changeLanguage}
          className='hidden min-w-8 px-2 font-mono text-[11px] font-semibold uppercase text-slate-500 sm:flex'
        >
          {language.toUpperCase()}
        </Button>
        <Dropdown placement='bottom-end'>
          <DropdownTrigger>
            <Button
              isIconOnly
              size='sm'
              radius='sm'
              variant='light'
              data-tour='controls'
              title={copy.workspaceMenu}
              className='text-slate-500'
            >
              {topologyActionPending ? <LoaderCircle className='h-4 w-4 animate-spin' /> : <Menu className='h-4 w-4' />}
            </Button>
          </DropdownTrigger>
          <DropdownMenu
            aria-label={copy.workspaceMenu}
            disabledKeys={!topologyActionsAvailable || topologyActionPending ? ['restart', 'wipe'] : []}
            onAction={(key) => {
              if (key === 'refresh') onRefresh();
              if (key === 'restart') onRestartTopology();
              if (key === 'wipe') onWipeTopology();
              if (key === 'tour') onStartTour();
            }}
          >
            <DropdownItem key='refresh' startContent={<RefreshCw className='h-4 w-4' />}>
              {copy.refresh}
            </DropdownItem>
            <DropdownItem key='restart' startContent={<RotateCcw className='h-4 w-4' />}>
              {copy.restartTopology}
            </DropdownItem>
            <DropdownItem
              key='wipe'
              color='danger'
              className='text-danger'
              startContent={<Eraser className='h-4 w-4' />}
            >
              {copy.wipeTopology}
            </DropdownItem>
            <DropdownItem key='tour' startContent={<GraduationCap className='h-4 w-4' />}>
              {copy.startTour}
            </DropdownItem>
          </DropdownMenu>
        </Dropdown>
        <Button
          isIconOnly
          size='sm'
          radius='sm'
          color='danger'
          variant='light'
          onPress={onStop}
          isDisabled={stopping}
          title={copy.stop}
        >
          {stopping ? <LoaderCircle className='h-4 w-4 animate-spin' /> : <Square className='h-4 w-4' />}
        </Button>
        <div
          className='ml-1 flex h-8 w-8 items-center justify-center rounded-full border border-slate-600 bg-slate-700 text-[11px] font-semibold text-slate-200'
          title={userName || session?.username}
        >
          {initials || <UserRound className='h-4 w-4' />}
        </div>
      </div>
    </header>
  );
};
