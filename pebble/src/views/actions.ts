import toast from 'react-hot-toast'
import { KeyedMutator, mutate as globalMutate } from 'swr'
import { SWRInfiniteKeyedMutator } from 'swr/infinite'

import { isAsyncFunction } from '@/utils/obj.ts'

import { Post, PostPagination } from '@/views/post/post-list.tsx'
import { Tag } from '@/views/tag/tag-list.tsx'

import {
  CLEAR_POSTS,
  CREATE_POST,
  DELETE_POST,
  DELETE_TAG,
  GET_OVERALL_COUNTS,
  GET_TAGS,
  RENAME_TAG,
  RESTORE_POST,
  STICK_TAG,
  UPDATE_POST,
  fetcher,
} from '@/api.ts'
import { AppError } from '@/error.ts'

export type ListMutator = SWRInfiniteKeyedMutator<PostPagination[]>
export type PostMutator = SWRInfiniteKeyedMutator<PostPagination[]> | KeyedMutator<Post>

interface TagActions {
  tagsMutator?: KeyedMutator<Tag[]>

  renameTag: (name: string, newName: string) => Promise<void>
  stickTag: (name: string, sticky: boolean) => Promise<void>
  deleteTag: (name: string) => Promise<void>

  refreshTags?: () => Promise<void>
}

interface PostActions {
  mainMutator?: ListMutator

  clearPosts: () => Promise<void>
  syncUpdateInMainPosts: (newPost: Partial<Post>) => Promise<void>
  syncDeleteInMainPosts: (id: number) => Promise<void>
  refreshMainPosts: () => Promise<void>

  createPost: (mutator: ListMutator, post: Partial<Post>) => Promise<void>
  updatePost: (mutator: PostMutator, post: Partial<Post>, revalidate?: boolean) => Promise<void>
  deletePost: (mutator: ListMutator, id: number, hard?: boolean) => Promise<void>
  restorePost: (mutator: ListMutator, id: number) => Promise<void>

  clearCaches: () => Promise<void>
}

export let postActions: PostActions = {
  createPost: async function (mutator, post) {
    await mutator(fetcher(CREATE_POST, post), {
      revalidate: true,
      populateCache: false,
    })
    await this.refreshMainPosts()
    await tagActions.refreshTags?.()
  },

  updatePost: async function (mutator, newPost, revalidate = true) {
    // @ts-expect-error make ts happy
    await mutator(fetcher(UPDATE_POST, newPost), {
      revalidate,
      rollbackOnError: true,
      populateCache: false,
      optimisticData: (pages: PostPagination[] | Post) => {
        if (Array.isArray(pages)) {
          return updateInPages(pages, newPost)
        } else {
          return { ...pages, ...newPost, updated_at: Date.now() }
        }
      },
    })

    if (newPost.content) {
      await tagActions.refreshTags?.()
    }

    if (location.pathname !== '/') {
      await this.syncUpdateInMainPosts(newPost)
    }
  },

  deletePost: async function (mutator, id, hard = false) {
    await mutator(fetcher(DELETE_POST, { id, hard }), {
      populateCache: false,
      revalidate: true,
      rollbackOnError: true,
      optimisticData: (pages) => {
        return deleteInPages(pages!, id)
      },
    })
    await tagActions.refreshTags?.()

    if (location.pathname !== '/') {
      await this.syncDeleteInMainPosts(id)
    }
  },

  restorePost: async function (mutator, id) {
    await mutator(fetcher(RESTORE_POST, { id }), {
      populateCache: false,
      revalidate: true,
      rollbackOnError: true,
      optimisticData: (pages) => {
        return deleteInPages(pages!, id)
      },
    })
    await tagActions.refreshTags?.()
  },

  clearPosts: async function () {
    await this.mainMutator?.(fetcher(CLEAR_POSTS, {}), {
      revalidate: true,
      populateCache: false,
    })
  },

  syncUpdateInMainPosts: async function (newPost) {
    await this.mainMutator?.(
      (pages) => {
        return updateInPages(pages!, newPost)
      },
      { revalidate: true },
    )
  },

  syncDeleteInMainPosts: async function (id) {
    await this.mainMutator?.(
      (pages) => {
        return deleteInPages(pages!, id)
      },
      { revalidate: true },
    )
  },

  refreshMainPosts: async function () {
    await this.mainMutator?.()
  },

  clearCaches: async function () {
    await globalMutate(() => true, undefined, { revalidate: false })
  },
}

export let tagActions: TagActions = {
  renameTag: async function (name, newName) {
    await this.tagsMutator?.(fetcher(RENAME_TAG, { name, new_name: newName }), {
      populateCache: false,
      revalidate: true,
    })

    await postActions.refreshMainPosts()
  },

  stickTag: async function (name, sticky) {
    await this.tagsMutator?.(fetcher(STICK_TAG, { name, sticky }), {
      populateCache: false,
      revalidate: false,
      rollbackOnError: true,
      optimisticData: (tags) => {
        return tags!.map((tag) => (tag.name === name ? { ...tag, sticky } : tag))
      },
    })
  },

  deleteTag: async function (name) {
    await this.tagsMutator?.(fetcher(DELETE_TAG, { name }), {
      populateCache: false,
      revalidate: false,
      rollbackOnError: true,
      optimisticData: (tags) => {
        return tags!.filter((tag) => tag.name !== name && !tag.name.startsWith(name + '/'))
      },
    })

    await globalMutate(GET_OVERALL_COUNTS)
    await postActions.refreshMainPosts()
  },

  refreshTags: async function () {
    await globalMutate(GET_TAGS)
  },
}

function updateInPages(pages: PostPagination[], newPost: Partial<Post>) {
  return pages.map((page) => ({
    ...page,
    posts: page.posts.map((post) =>
      post.id === newPost.id ? { ...post, ...newPost, updated_at: Date.now() } : post,
    ),
  }))
}

function deleteInPages(pages: PostPagination[], id: number) {
  return pages.map((page) => ({
    ...page,
    posts: page.posts.filter((post) => post.id !== id),
  }))
}

function handleAsyncError<T extends object>(target: T): T {
  return new Proxy(target, {
    get(target, prop, receiver) {
      const value = Reflect.get(target, prop, receiver)
      if (!isAsyncFunction(value)) return value

      return async (...args: unknown[]) => {
        try {
          return await value.apply(target, args)
        } catch (err) {
          if (err instanceof AppError) {
            toast.error(err.friendlyMessage, { id: 'AppError' })
          } else {
            toast.error('Server Error, please try again later')
          }
          throw err
        }
      }
    },
  })
}

postActions = handleAsyncError(postActions)
tagActions = handleAsyncError(tagActions)
