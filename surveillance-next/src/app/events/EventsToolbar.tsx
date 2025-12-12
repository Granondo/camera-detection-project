'use client'

import { HStack, chakra } from '@chakra-ui/react'
import { useRouter, usePathname, useSearchParams } from 'next/navigation'
import { useCallback } from 'react'

const EVENT_TYPES = [
  'motion_detected',
  'object_detected',
  'sound_detected',
  'camera_offline',
  'camera_online',
]

export function EventsToolbar() {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const type = searchParams.get('type')

  const createQueryString = useCallback(
    (name: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString())
      if (value) {
        params.set(name, value)
      } else {
        params.delete(name)
      }
      // Reset page to 1 when filter changes
      params.set('page', '1')
      return params.toString()
    },
    [searchParams]
  )

  const handleTypeChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    router.push(pathname + '?' + createQueryString('type', event.target.value))
  }

  return (
    <HStack mb={6} justify="flex-end">
      <chakra.select
        maxW="200px"
        value={type || ''}
        onChange={handleTypeChange}
        p={2}
        borderWidth="1px"
        borderRadius="md"
        bg="white"
      >
        <option value="">Filter by type</option>
        {EVENT_TYPES.map((type) => (
          <option key={type} value={type}>
            {type.replace(/_/g, ' ')}
          </option>
        ))}
      </chakra.select>
    </HStack>
  )
}
