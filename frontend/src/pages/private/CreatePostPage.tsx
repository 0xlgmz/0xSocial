import { AuthenticatedLayout } from '../../components/layout/AuthenticatedLayout'
import { PostComposer } from '../../components/post/PostComposer'
import { useAuth } from '../../features/auth/useAuth'
import { navigate } from '../../lib/routes'

export function CreatePostPage() {
  const { profile, refreshProfile } = useAuth()
  if (!profile) return null

  return (
    <AuthenticatedLayout title="Create post" route="/create">
      <div className="flex flex-col gap-4">
        <div className="px-1">
          <h1 className="text-2xl font-semibold tracking-[-.03em]">Create a post</h1>
          <p className="mt-1 text-sm text-base-content/50">Share an update with your network.</p>
        </div>
        <PostComposer
          displayName={profile.displayName}
          handle={profile.handle}
          avatarUrl={profile.avatarUrl}
          onCreated={() => {
            void refreshProfile()
            navigate('/feed')
          }}
        />
      </div>
    </AuthenticatedLayout>
  )
}
