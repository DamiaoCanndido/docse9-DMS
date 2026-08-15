import { getMe } from '@/app/api/auth';
import { ProfileContent } from '@/components/ProfileContent';

export const dynamic = 'force-dynamic';

export default async function ProfilePage() {
  const currentUser = await getMe();

  return (
    <div className="w-full flex-1">
      <ProfileContent currentUser={currentUser} />
    </div>
  );
}
