import { Navigate, useParams } from 'react-router-dom';
import { RoutesLocation } from '@/components/routes';

export default function LegacyTopologyPage() {
  const { sessionId } = useParams();
  return <Navigate replace to={RoutesLocation.sessionTopology(sessionId || '')} />;
}
