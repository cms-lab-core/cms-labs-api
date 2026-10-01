import { useCallback, useEffect, useMemo, useState } from 'react';
import { createPortal } from 'react-dom';
import { ArrowLeft, ArrowRight, Check, X } from 'lucide-react';
import { Button } from '@heroui/react';
import type { LabCopy } from './copy';

interface WorkspaceTourProps {
  copy: LabCopy;
  open: boolean;
  onComplete: () => void;
}

interface TourStep {
  title: string;
  description: string;
  selector?: string;
}

interface TargetRect {
  top: number;
  left: number;
  right: number;
  bottom: number;
  width: number;
  height: number;
}

const tourPadding = 8;
const tourPanelHeight = 320;
const tourViewportMargin = 16;

export const WorkspaceTour = ({ copy, open, onComplete }: WorkspaceTourProps) => {
  const [stepIndex, setStepIndex] = useState(0);
  const [target, setTarget] = useState<TargetRect>();
  const steps = useMemo<TourStep[]>(
    () => [
      { title: copy.tourWelcomeTitle, description: copy.tourWelcomeDescription },
      { title: copy.tourTabsTitle, description: copy.tourTabsDescription, selector: '[data-tour="tabs"]' },
      {
        title: copy.tourWorkspaceTitle,
        description: copy.tourWorkspaceDescription,
        selector: '[data-tour="workspace"]'
      },
      { title: copy.tourChecksTitle, description: copy.tourChecksDescription, selector: '[data-tour="checks"]' },
      {
        title: copy.tourControlsTitle,
        description: copy.tourControlsDescription,
        selector: '[data-tour="controls"]'
      },
      {
        title: copy.tourAssignmentTitle,
        description: copy.tourAssignmentDescription,
        selector: '[data-workspace-tab="assignment"]'
      }
    ],
    [copy]
  );

  const updateTarget = useCallback(() => {
    const selector = steps[stepIndex]?.selector;
    if (!selector) {
      setTarget(undefined);
      return;
    }
    const element = document.querySelector<HTMLElement>(selector);
    if (!element) {
      setTarget(undefined);
      return;
    }
    const rect = element.getBoundingClientRect();
    setTarget({
      top: Math.max(0, rect.top - tourPadding),
      left: Math.max(0, rect.left - tourPadding),
      right: Math.min(window.innerWidth, rect.right + tourPadding),
      bottom: Math.min(window.innerHeight, rect.bottom + tourPadding),
      width: rect.width + tourPadding * 2,
      height: rect.height + tourPadding * 2
    });
  }, [stepIndex, steps]);

  useEffect(() => {
    if (!open) return;
    setStepIndex(0);
  }, [open]);

  useEffect(() => {
    if (!open) return;
    updateTarget();
    window.addEventListener('resize', updateTarget);
    window.addEventListener('scroll', updateTarget, true);
    return () => {
      window.removeEventListener('resize', updateTarget);
      window.removeEventListener('scroll', updateTarget, true);
    };
  }, [open, updateTarget]);

  if (!open) return null;
  const step = steps[stepIndex];
  const last = stepIndex === steps.length - 1;
  const panelStyle = target
    ? (() => {
        const below = target.bottom + 14;
        const above = target.top - tourPanelHeight - 14;
        const maxTop = Math.max(tourViewportMargin, window.innerHeight - tourPanelHeight - tourViewportMargin);
        const top =
          below + tourPanelHeight <= window.innerHeight - tourViewportMargin
            ? below
            : above >= tourViewportMargin
              ? above
              : Math.min(maxTop, Math.max(tourViewportMargin, (window.innerHeight - tourPanelHeight) / 2));
        return {
          top,
          left: Math.max(tourViewportMargin, Math.min(target.left, window.innerWidth - 376))
        };
      })()
    : undefined;

  return createPortal(
    <div className='fixed inset-0 z-[10000]' role='dialog' aria-modal='true' aria-label={step.title}>
      {target ? (
        <>
          <div className='fixed left-0 right-0 top-0 bg-black/65' style={{ height: target.top }} />
          <div className='fixed bottom-0 left-0 right-0 bg-black/65' style={{ top: target.bottom }} />
          <div
            className='fixed left-0 bg-black/65'
            style={{ top: target.top, width: target.left, height: target.height }}
          />
          <div
            className='fixed right-0 bg-black/65'
            style={{ top: target.top, left: target.right, height: target.height }}
          />
          <div
            className='pointer-events-none fixed rounded-lg border-2 border-emerald-400 shadow-[0_0_0_4px_rgb(16_185_129_/_0.18),0_0_28px_rgb(16_185_129_/_0.35)]'
            style={{ top: target.top, left: target.left, width: target.width, height: target.height }}
          />
        </>
      ) : (
        <div className='fixed inset-0 bg-black/65' />
      )}

      <div
        className={`fixed max-h-[calc(100vh-32px)] w-[min(360px,calc(100vw-32px))] overflow-y-auto rounded-lg border border-slate-200 bg-white p-5 text-slate-900 shadow-2xl dark:border-slate-700 dark:bg-[#182235] dark:text-slate-100 ${
          target ? '' : 'left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2'
        }`}
        style={panelStyle}
      >
        <Button
          isIconOnly
          size='sm'
          variant='light'
          onPress={onComplete}
          className='absolute right-2 top-2 text-slate-400'
          aria-label={copy.tourSkip}
        >
          <X className='h-4 w-4' />
        </Button>
        <p className='text-[11px] font-semibold uppercase tracking-[0.16em] text-emerald-600 dark:text-emerald-400'>
          {copy.tourStep} {stepIndex + 1} / {steps.length}
        </p>
        <h2 className='mt-2 pr-6 text-lg font-semibold'>{step.title}</h2>
        <p className='mt-2 text-sm leading-6 text-slate-600 dark:text-slate-400'>{step.description}</p>
        <div className='mt-5 flex items-center justify-between gap-3'>
          <Button size='sm' variant='light' onPress={onComplete} className='text-xs text-slate-500'>
            {copy.tourSkip}
          </Button>
          <div className='flex gap-2'>
            {stepIndex > 0 && (
              <Button
                size='sm'
                variant='bordered'
                onPress={() => setStepIndex((value) => value - 1)}
                startContent={<ArrowLeft className='h-3.5 w-3.5' />}
              >
                {copy.tourPrevious}
              </Button>
            )}
            <Button
              size='sm'
              color='success'
              onPress={() => (last ? onComplete() : setStepIndex((value) => value + 1))}
              startContent={last ? <Check className='h-3.5 w-3.5' /> : undefined}
              endContent={!last ? <ArrowRight className='h-3.5 w-3.5' /> : undefined}
            >
              {last ? copy.tourDone : copy.tourNext}
            </Button>
          </div>
        </div>
      </div>
    </div>,
    document.body
  );
};
