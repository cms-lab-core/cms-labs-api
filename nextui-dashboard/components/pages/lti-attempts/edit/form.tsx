'use client';
import React from 'react';
import { Accordion, AccordionItem, addToast, Button, Input, Select, SelectItem } from '@heroui/react';
import { Formik } from 'formik';
import { ModelsLTIAttempt } from '@/helpers/api';
import useLanguageBrowser from '@/helpers/locale';
import { useNavigate } from 'react-router-dom';
import { RoutesLocation } from '@/components/routes';
import dayjs from 'dayjs';
import { Loading } from '@/components/scroll/loader';
import { useConfirmPopup } from '@/components/hooks/useDeletePopup';
import { ServerInput } from '@/components/pages/lti-attempts/edit/autoCompleteServer';
import { CamelCasedPropertiesDeep } from 'type-fest';
import { useQueryLtiAttemptGet } from '@/helpers/queries/lti_attempt/use-query-lti-attempt-get';
import { useQueryUserGet } from '@/helpers/queries/user/use-query-user-get';
import { useQueryLtiRoutingGet } from '@/helpers/queries/lti_routing/use-query-lti-routing-get';
import { useMutationLtiAttemptUpdate } from '@/helpers/queries/lti_attempt/use-mutation-lti-attempt-update';
import { useMutationLtiAttemptDelete } from '@/helpers/queries/lti_attempt/use-mutation-lti-attempt-delete';
import { LogIn } from 'lucide-react';

interface EditFormProps {
  id?: number;
}

const defaultValues: CamelCasedPropertiesDeep<ModelsLTIAttempt> = {
  attemptId: '',
  createdAt: '',
  ltiRoutingId: 0,
  serverId: 0,
  roomId: 0,
  userId: 0,
  updatedAt: '',
  status: '',
  result: undefined,
  synchronizedAt: ''
};

export const LtiAttemptEditForm = ({ id }: EditFormProps) => {
  const {
    locale: {
      AuthProviderAttempt,
      Forms,
      Sidebar,
      Tables: { LTIAttemptsTable }
    }
  } = useLanguageBrowser();
  const navigate = useNavigate();
  const response = useQueryLtiAttemptGet({ id });
  const initialValues = response.data?.model ?? defaultValues;
  const responseUser = useQueryUserGet({ id: initialValues.userId });
  const user = responseUser.data?.model;
  const responseRoute = useQueryLtiRoutingGet({ id: initialValues.ltiRoutingId });
  const route = responseRoute.data?.model;
  const isLoading = response.isLoading || responseUser.isLoading || responseRoute.isLoading;
  const { mutate } = useMutationLtiAttemptUpdate({
    onSuccess: (data) => {
      navigate(RoutesLocation.ltiAttemptEdit(data?.id?.toString() || ''), { replace: true });
      addToast({
        title: Forms.SaveSuccess,
        color: 'success'
      });
    },
    onError: (error: any) => {
      addToast({
        title: Forms.SaveError,
        description: error.data.message,
        color: 'danger'
      });
    }
  });
  const onDeleteMutation = useMutationLtiAttemptDelete({
    onSuccess: () => {
      navigate(RoutesLocation.ltiAttempts({ userId: initialValues.userId }), { replace: true });
      addToast({
        title: Forms.DeleteSuccess,
        color: 'success'
      });
    },
    onError: (error: any) => {
      addToast({
        title: Forms.DeleteError,
        description: error.data.message,
        color: 'danger'
      });
    }
  });
  const ltiAttemptDeletePopup = useConfirmPopup({
    title: AuthProviderAttempt.DeletePopup.Title,
    description: AuthProviderAttempt.DeletePopup.Description,
    onConfirm: onDeleteMutation.mutate.bind(onDeleteMutation.mutate, { id })
  });
  if (isLoading) return <Loading size='md' />;
  return (
    <Formik
      initialValues={initialValues}
      validationSchema={undefined}
      onSubmit={(values, formikHelpers) => {
        mutate({
          id: values.id,
          serverId: parseInt(values.serverId?.toString() || ''),
          status: values.status,
          result: values.result
        });
      }}
    >
      {({ values, handleChange, setFieldValue, handleSubmit }) => (
        <>
          {ltiAttemptDeletePopup.component({})}
          <div className='flex flex-col gap-4 mb-4'>
            <Input
              variant='bordered'
              label={AuthProviderAttempt.FieldID}
              value={(initialValues.attemptId ?? '').toString()}
              isReadOnly
            />
            <ServerInput
              variant='bordered'
              label={AuthProviderAttempt.FieldPNETServer}
              value={(values.serverId ?? '').toString()}
              onChange={handleChange('serverId')}
            />
            <Select
              variant='bordered'
              label={AuthProviderAttempt.FieldStatus}
              selectedKeys={[values.status || '']}
              onChange={handleChange('status')}
            >
              {AuthProviderAttempt.FieldStatusValues.map(({ key, value, description }) => (
                <SelectItem key={key} description={description}>
                  {value}
                </SelectItem>
              ))}
            </Select>
            <Accordion>
              <AccordionItem
                key='grade'
                aria-label={AuthProviderAttempt.FieldResult}
                title={AuthProviderAttempt.FieldResult}
              >
                <div className='flex flex-col gap-4'>
                  <Input
                    variant='bordered'
                    label={AuthProviderAttempt.FieldResultMaxScore}
                    type='number'
                    value={
                      values.result && typeof values.result === 'object'
                        ? (values.result as any).maxScore?.toString() || ''
                        : ''
                    }
                    onChange={(e) => {
                      const newResult = {
                        ...(values.result as object),
                        maxScore: parseFloat(e.target.value) || 0,
                        currentScore: (values.result as any)?.currentScore || 0,
                        resultDisplay: (values.result as any)?.resultDisplay || ''
                      };
                      setFieldValue('result', newResult);
                    }}
                  />
                  <Input
                    variant='bordered'
                    label={AuthProviderAttempt.FieldResultCurrentScore}
                    type='number'
                    value={
                      values.result && typeof values.result === 'object'
                        ? (values.result as any).currentScore?.toString() || ''
                        : ''
                    }
                    onChange={(e) => {
                      const newResult = {
                        ...(values.result as object),
                        maxScore: (values.result as any)?.maxScore || 0,
                        currentScore: parseFloat(e.target.value) || 0,
                        resultDisplay: (values.result as any)?.resultDisplay || ''
                      };
                      setFieldValue('result', newResult);
                    }}
                  />
                  <Input
                    variant='bordered'
                    label={AuthProviderAttempt.FieldResultDisplay}
                    type='text'
                    value={
                      values.result && typeof values.result === 'object'
                        ? (values.result as any).resultDisplay?.toString() || ''
                        : ''
                    }
                    onChange={(e) => {
                      const newResult = {
                        ...(values.result as object),
                        maxScore: (values.result as any)?.maxScore || 0,
                        currentScore: (values.result as any)?.currentScore || 0,
                        resultDisplay: e.target.value
                      };
                      setFieldValue('result', newResult);
                    }}
                  />
                </div>
              </AccordionItem>
            </Accordion>
            {initialValues.roomId && (
              <Input
                variant='bordered'
                label={AuthProviderAttempt.FieldRoomNumber}
                type='number'
                value={(initialValues.roomId ?? 0).toString()}
                isReadOnly
              />
            )}
            <Input
              variant='bordered'
              label={AuthProviderAttempt.FieldUserId}
              type='number'
              value={(initialValues.userId ?? 0).toString()}
              isReadOnly
            />
            <Input
              variant='bordered'
              label={AuthProviderAttempt.FieldUserName}
              type='text'
              value={user?.name ?? ''}
              isReadOnly
            />
            <Input
              variant='bordered'
              label={AuthProviderAttempt.FieldUserEmail}
              type='text'
              value={user?.email ?? ''}
              isReadOnly
            />
            <Input
              variant='bordered'
              label={AuthProviderAttempt.FieldLTIRoutingID}
              type='number'
              value={(initialValues.ltiRoutingId ?? 0).toString()}
              isReadOnly
            />
            <Input
              variant='bordered'
              label={AuthProviderAttempt.FieldLTIRoutingName}
              type='text'
              value={route?.name ?? ''}
              isReadOnly
            />
            <Input
              variant='bordered'
              label={AuthProviderAttempt.FieldSynchronizedAt}
              type='datetime-local'
              value={dayjs(initialValues.synchronizedAt ?? '').format('YYYY-MM-DDTHH:mm')}
              isReadOnly
            />
            <Input
              variant='bordered'
              label={AuthProviderAttempt.FieldCreatedAt}
              type='datetime-local'
              value={dayjs(initialValues.createdAt ?? '').format('YYYY-MM-DDTHH:mm')}
              isReadOnly
            />
            <Input
              variant='bordered'
              label={AuthProviderAttempt.FieldUpdatedAt}
              type='datetime-local'
              value={dayjs(initialValues.updatedAt ?? '').format('YYYY-MM-DDTHH:mm')}
              isReadOnly
            />
            <Button onPress={() => handleSubmit()} variant='flat' color='primary'>
              {Sidebar.Save}
            </Button>
            {(initialValues.status === 'pending' || initialValues.status === 'active') && initialValues.attemptId && (
              <Button
                onPress={() => navigate(RoutesLocation.session(initialValues.attemptId))}
                variant='flat'
                color='success'
                startContent={<LogIn className='h-4 w-4' />}
              >
                {LTIAttemptsTable.Connect}
              </Button>
            )}
            <Button onPress={ltiAttemptDeletePopup.onOpen} variant='flat' color='danger'>
              {Sidebar.Delete}
            </Button>
          </div>
        </>
      )}
    </Formik>
  );
};
