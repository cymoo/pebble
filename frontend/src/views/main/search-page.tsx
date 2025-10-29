import { ArrowDownUpIcon, SearchIcon } from 'lucide-react'
import { useRef } from 'react'
import { useLocation, useSearchParams } from 'react-router'

import { cx } from '@/utils/css.ts'
import { debounce } from '@/utils/func.ts'
import { useLatest } from '@/utils/hooks/use-latest.ts'

import { Button } from '@/components/button.tsx'
import { Input } from '@/components/input.tsx'
import { T } from '@/components/translation.tsx'

import { OrderBy, PostList } from '@/views/post/post-list.tsx'

export function SearchPage() {
  const [params, setParams] = useSearchParams()

  const searchTerm = params.get('query') || ''
  const orderBy = (params.get('order_by') as OrderBy | null) ?? undefined
  const asc = params.get('asc') == 'true'

  const location = useLocation()
  const prevState = location.state as unknown

  const latestParams = useLatest(params)

  const handleQueryChange = useRef(
    debounce((query: string) => {
      const params = latestParams.current
      params.set('query', query)
      setParams(params, { replace: true, state: prevState })
    }, 500),
  ).current

  const handleOrderChange = (orderBy: string, asc = false) => {
    params.set('order_by', orderBy)
    params.set('asc', String(asc))
    setParams(params, { replace: true, state: prevState })
  }

  return (
    <div className="flex vh-full flex-col p-3">
      <Input
        className="focus-visible:ring-transparent flex-none"
        autoFocus={true}
        prefix={<SearchIcon className="size-4" />}
        type="search"
        placeholder="Search..."
        defaultValue={searchTerm}
        onChange={(event) => {
          handleQueryChange(event.target.value.trim())
        }}
      />
      <div className="my-1 flex flex-none items-center justify-end opacity-80 *:text-xs">
        <ArrowDownUpIcon className="mr-1 size-4" aria-hidden="true" />
        <Button
          className={cx({ 'text-primary': orderBy === 'score' || orderBy === undefined })}
          size="sm"
          variant="ghost"
          onClick={() => {
            handleOrderChange('score', false)
          }}
        >
          <T name="relevance" />
        </Button>
        <Button
          className={cx({ 'text-primary': orderBy === 'created_at' && !asc })}
          size="sm"
          variant="ghost"
          onClick={() => {
            handleOrderChange('created_at', false)
          }}
        >
          <T name="newest" />
        </Button>
        <Button
          className={cx({ 'text-primary': orderBy === 'created_at' && asc }, 'mr-2')}
          size="sm"
          variant="ghost"
          onClick={() => {
            handleOrderChange('created_at', true)
          }}
        >
          <T name="oldest" />
        </Button>
      </div>
      {searchTerm && (
        <PostList
          className="flex-1"
          queryString={`query=${searchTerm}`}
          orderBy={orderBy}
          ascending={asc}
        />
      )}
    </div>
  )
}
