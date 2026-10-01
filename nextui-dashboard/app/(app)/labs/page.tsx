'use client';

import { useMemo, useState } from 'react';
import { Button, Card, CardBody, Chip, Input, Spinner } from '@heroui/react';
import { BookOpen, GitBranch, Network, Play, Search, Users } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { Layout } from '@/components/layout/layout';
import { RoutesLocation } from '@/components/routes';
import useLanguageBrowser from '@/helpers/locale';
import { useMutationLabCatalogStart } from '@/helpers/queries/lab_catalog/use-mutation-lab-catalog-start';
import { useIsLabCatalogVisible } from '@/helpers/queries/lab_catalog/use-is-lab-catalog-visible';
import { useQueryLabCatalog } from '@/helpers/queries/lab_catalog/use-query-lab-catalog';

const LabCatalogPage = () => {
  const navigate = useNavigate();
  const { locale } = useLanguageBrowser();
  const copy = locale.LabCatalog;
  const [search, setSearch] = useState('');
  const query = useQueryLabCatalog(search.trim());
  const start = useMutationLabCatalogStart();
  const labs = useMemo(() => query.data?.labs ?? [], [query.data?.labs]);
  const isVisible = useIsLabCatalogVisible();

  const openLab = (labId: number, attemptId?: string) => {
    if (attemptId) {
      navigate(RoutesLocation.session(attemptId));
      return;
    }
    start.mutate(
      { labId },
      {
        onSuccess: (result) => navigate(result.nextUrl || RoutesLocation.session(result.attemptId))
      }
    );
  };

  if (!isVisible) {
    return null;
  }

  return (
    <Layout>
      <main className='min-h-screen bg-background px-6 py-8 text-foreground lg:px-10'>
        <div className='mx-auto flex w-full max-w-6xl flex-col gap-6'>
          <header className='flex flex-col gap-2'>
            <div className='flex items-center gap-3'>
              <div className='rounded-lg bg-primary/10 p-2 text-primary'>
                <BookOpen className='h-6 w-6' />
              </div>
              <div>
                <h1 className='text-2xl font-semibold'>{copy.Title}</h1>
                <p className='text-sm text-default-500'>{copy.Description}</p>
              </div>
            </div>
            <Input
              className='mt-3 max-w-xl'
              value={search}
              onValueChange={setSearch}
              placeholder={copy.Search}
              startContent={<Search className='h-4 w-4 text-default-400' />}
              isClearable
              onClear={() => setSearch('')}
            />
          </header>

          {query.isLoading && (
            <div className='flex justify-center py-20'>
              <Spinner label={copy.Loading} />
            </div>
          )}
          {query.isError && (
            <Card className='border border-danger/30 bg-danger/5'>
              <CardBody className='flex flex-row items-center justify-between gap-4'>
                <span>{copy.Error}</span>
                <Button color='danger' variant='flat' onPress={() => query.refetch()}>
                  {copy.Retry}
                </Button>
              </CardBody>
            </Card>
          )}
          {!query.isLoading && !query.isError && labs.length === 0 && (
            <div className='rounded-xl border border-dashed border-default-300 px-6 py-20 text-center text-default-500'>
              {copy.Empty}
            </div>
          )}

          <section className='grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3'>
            {labs.map((lab) => {
              const isStarting = start.isPending && start.variables?.labId === lab.id;
              return (
                <Card key={lab.id} className='border border-default-200 bg-content1 shadow-sm'>
                  <CardBody className='flex min-h-64 flex-col gap-4 p-5'>
                    <div className='flex items-start justify-between gap-3'>
                      <div className='rounded-md bg-secondary/10 p-2 text-secondary'>
                        <Network className='h-5 w-5' />
                      </div>
                      {lab.attempt && (
                        <Chip color={lab.attempt.status === 'active' ? 'success' : 'warning'} size='sm' variant='flat'>
                          {lab.attempt.status}
                        </Chip>
                      )}
                    </div>
                    <div className='flex-1'>
                      <h2 className='text-lg font-semibold'>{lab.name}</h2>
                      <p className='mt-2 line-clamp-4 text-sm text-default-500'>
                        {lab.description || copy.NoDescription}
                      </p>
                    </div>
                    <div className='flex flex-col gap-1 text-xs text-default-500'>
                      <span className='flex items-center gap-2'>
                        <Users className='h-3.5 w-3.5' />
                        {copy.Collaboration}: {lab.collaboration}
                      </span>
                      <span className='flex min-w-0 items-center gap-2'>
                        <GitBranch className='h-3.5 w-3.5 shrink-0' />
                        <span className='truncate'>{lab.repository}</span>
                      </span>
                    </div>
                    <Button
                      color='primary'
                      startContent={!isStarting && <Play className='h-4 w-4' />}
                      isLoading={isStarting}
                      onPress={() => openLab(lab.id, lab.attempt?.id)}
                    >
                      {lab.attempt ? copy.Continue : copy.Start}
                    </Button>
                  </CardBody>
                </Card>
              );
            })}
          </section>
          {start.isError && <p className='text-sm text-danger'>{copy.StartError}</p>}
        </div>
      </main>
    </Layout>
  );
};

export default LabCatalogPage;
