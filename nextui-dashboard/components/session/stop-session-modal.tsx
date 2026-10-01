import { useEffect, useState } from 'react';
import { AlertTriangle, Trash2 } from 'lucide-react';
import { Button, Checkbox, Modal, ModalBody, ModalContent, ModalFooter, ModalHeader } from '@heroui/react';
import type { LabCopy } from './copy';

interface StopSessionModalProps {
  copy: LabCopy;
  open: boolean;
  stopping: boolean;
  error?: string;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
}

export const StopSessionModal = ({ copy, open, stopping, error, onOpenChange, onConfirm }: StopSessionModalProps) => {
  const [teacherChecked, setTeacherChecked] = useState(false);
  const [filesExported, setFilesExported] = useState(false);
  const [irreversibleAccepted, setIrreversibleAccepted] = useState(false);

  useEffect(() => {
    if (!open) {
      setTeacherChecked(false);
      setFilesExported(false);
      setIrreversibleAccepted(false);
    }
  }, [open]);

  const confirmed = teacherChecked && filesExported && irreversibleAccepted;

  return (
    <Modal
      isOpen={open}
      onOpenChange={onOpenChange}
      isDismissable={!stopping}
      hideCloseButton={stopping}
      placement='center'
      backdrop='blur'
      classNames={{
        base: 'border border-danger-200 dark:border-danger-900',
        backdrop: 'z-[9000]',
        wrapper: 'z-[9000]'
      }}
    >
      <ModalContent>
        <ModalHeader className='flex items-start gap-3'>
          <span className='mt-0.5 rounded-full bg-danger-50 p-2 text-danger dark:bg-danger-950'>
            <AlertTriangle className='h-5 w-5' />
          </span>
          <span>
            <span className='block text-lg'>{copy.stopTitle}</span>
            <span className='mt-1 block text-xs font-medium uppercase tracking-wider text-danger'>
              {copy.stopWarning}
            </span>
          </span>
        </ModalHeader>
        <ModalBody>
          <p className='text-sm leading-6 text-default-500'>{copy.stopDescription}</p>
          <div className='mt-2 flex flex-col gap-4 rounded-medium border border-divider bg-content2/50 p-4'>
            <Checkbox isSelected={teacherChecked} onValueChange={setTeacherChecked} color='warning'>
              <span className='text-sm leading-5'>{copy.stopTeacherChecked}</span>
            </Checkbox>
            <Checkbox isSelected={filesExported} onValueChange={setFilesExported} color='warning'>
              <span className='text-sm leading-5'>{copy.stopFilesExported}</span>
            </Checkbox>
            <Checkbox isSelected={irreversibleAccepted} onValueChange={setIrreversibleAccepted} color='danger'>
              <span className='text-sm font-medium leading-5 text-danger'>{copy.stopIrreversible}</span>
            </Checkbox>
          </div>
          {error && (
            <p className='rounded-medium bg-danger-50 px-3 py-2 text-sm text-danger dark:bg-danger-950'>{error}</p>
          )}
        </ModalBody>
        <ModalFooter>
          <Button variant='flat' onPress={() => onOpenChange(false)} isDisabled={stopping}>
            {copy.stopCancel}
          </Button>
          <Button
            color='danger'
            startContent={!stopping ? <Trash2 className='h-4 w-4' /> : undefined}
            isDisabled={!confirmed || stopping}
            isLoading={stopping}
            onPress={onConfirm}
          >
            {copy.stopConfirm}
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
