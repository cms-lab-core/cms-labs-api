'use client';
import { LanguageType } from '@/helpers/locale/locale';
import { Button, Image } from '@heroui/react';
import FlagEn from './flagEn.svg';
import FlagRu from './flagRu.svg';
import useLanguageBrowser from '@/helpers/locale';
import { useNavigate } from 'react-router-dom';
import { Modal, ModalBody, ModalContent, ModalHeader } from '@heroui/modal';
import React from 'react';

const languages = [
  {
    text: 'Русский',
    img: FlagRu,
    lang: 'ru' as LanguageType
  },
  {
    text: 'English',
    img: FlagEn,
    lang: 'en' as LanguageType
  }
];

const LanguageSwitcher = () => {
  const navigate = useNavigate();
  const { locale, setLang } = useLanguageBrowser();
  const handleClose = () => {
    navigate(-1);
  };

  const langBlock = languages.map((item) => (
    <Button
      key={item.lang}
      variant='bordered'
      onPress={() => {
        setLang(item.lang);
        navigate(-1);
      }}
      className='h-auto flex-1 flex-col gap-2 py-5'
    >
      <Image src={item.img} alt={item.text} width='64' height='42' className='h-10 w-16 object-cover' />
      <span>{item.text}</span>
    </Button>
  ));

  return (
    <Modal isOpen={true} onClose={handleClose}>
      <ModalContent>
        <ModalHeader className='flex flex-col gap-1'>{locale.LanguageSwitcher.LanguageSwitch}</ModalHeader>
        <ModalBody className='flex-row gap-3 pb-6'>{langBlock}</ModalBody>
      </ModalContent>
    </Modal>
  );
};

export default LanguageSwitcher;
