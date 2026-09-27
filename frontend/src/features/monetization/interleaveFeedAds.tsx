import type { ReactNode } from 'react'
import type { FeedAdPlacement } from '../../api/monetization'

export function interleaveFeedAds<T>(
  posts: readonly T[],
  placement: FeedAdPlacement,
  renderPost: (post: T) => ReactNode,
  renderAd: (adIndex: number) => ReactNode,
): ReactNode[] {
  if (
    posts.length === 0
    || !placement.enabled
    || !Number.isSafeInteger(placement.firstAfterPosts)
    || !Number.isSafeInteger(placement.repeatEveryPosts)
    || !Number.isSafeInteger(placement.maximumAds)
    || placement.firstAfterPosts <= 0
    || placement.repeatEveryPosts <= 0
    || placement.maximumAds <= 0
  ) return posts.map(renderPost)

  const output: ReactNode[] = []
  let adCount = 0
  let nextAdAfter = placement.firstAfterPosts

  posts.forEach((post, index) => {
    output.push(renderPost(post))
    const renderedPosts = index + 1
    if (adCount < placement.maximumAds && renderedPosts === nextAdAfter) {
      output.push(renderAd(adCount))
      adCount += 1
      nextAdAfter += placement.repeatEveryPosts
    }
  })

  return output
}
