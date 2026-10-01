import React from 'react';
import { Button, useDisclosure } from '@heroui/react';
import useLanguageBrowser from '@/helpers/locale';
import { Modal, ModalBody, ModalContent, ModalFooter, ModalHeader } from '@heroui/modal';
import { zIndexClassModal } from '@/components/providers/const';

interface UseConfirmPopupProps {
  title?: string;
  description?: string;
  onConfirm?: () => void;
}

export const useConfirmPopup = (params: UseConfirmPopupProps) => {
  const { isOpen, onOpen, onOpenChange } = useDisclosure();
  const {
    locale: { Sidebar }
  } = useLanguageBrowser();
  const component = (props: UseConfirmPopupProps) => {
    return (
      <Modal
        classNames={{
          wrapper: zIndexClassModal,
          backdrop: zIndexClassModal
        }}
        isOpen={isOpen}
        onOpenChange={onOpenChange}
      >
        <ModalContent>
          {(onClose) => (
            <>
              <ModalHeader className='flex flex-col gap-1'>{props.title || params.title || 'Confirm'}</ModalHeader>
              <ModalBody>
                <p>{props.description || params.description || ''}</p>
              </ModalBody>
              <ModalFooter>
                <Button color='danger' variant='light' onPress={onClose}>
                  {Sidebar.Close}
                </Button>
                <Button
                  color='primary'
                  onPress={() => {
                    (props.onConfirm || params.onConfirm)?.();
                    onClose();
                  }}
                >
                  {Sidebar.Confirm}
                </Button>
              </ModalFooter>
            </>
          )}
        </ModalContent>
      </Modal>
    );
  };
  return { component, onOpen };
};
