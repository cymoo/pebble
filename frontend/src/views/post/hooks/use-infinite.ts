import useSWRInfinite from 'swr/infinite'

import { GET_POSTS, SEARCH, fetcher } from '@/api.ts'
import { AppError } from '@/error.ts'

import { Post, PostPagination } from '../post-list.tsx'

export const useInfinitePosts = (queryString?: string) => {
  const { data, error, isLoading, isValidating, mutate, size, setSize } = useSWRInfinite<
    PostPagination,
    AppError
  >(
    (_pageIndex, previousPageData: PostPagination | null) => {
      // reached the end
      if (previousPageData?.cursor === -1) {
        return null
      }

      const params = new URLSearchParams(queryString)
      if (params.has('query')) {
        return `${SEARCH}?${params.toString()}&partial=true`
      }

      if (previousPageData) params.set('cursor', previousPageData.cursor.toString())
      return `${GET_POSTS}?${params.toString()}`
    },
    fetcher as (url: string) => Promise<PostPagination>,
    {
      initialSize: 1,
      persistSize: false,
    },
  )

  const posts = data ? ([] as Post[]).concat(...data.map((item) => item.posts)) : []
  const isLoadingMore =
    isLoading || (isValidating && size > 0 && data && typeof data[size - 1] === 'undefined')
  const isEmpty = data?.[0]?.posts?.length === 0
  const isReachingEnd = isEmpty || (data && data[data.length - 1]?.cursor === -1)
  const isRefreshing = isValidating && data && data.length === size

  return {
    posts,
    pagesData: data,
    error,
    isLoadingMore,
    isEmpty,
    isRefreshing,
    isReachingEnd,
    mutate,
    setSize,
  }
}
