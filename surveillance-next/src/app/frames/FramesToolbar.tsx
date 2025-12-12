'use client'

import { HStack, chakra } from '@chakra-ui/react'
import { useRouter, usePathname, useSearchParams } from 'next/navigation'
import { useCallback } from 'react'

// In a real app, you might fetch these from an API
const CAMERA_IDS = [1, 2, 3, 4]

export function FramesToolbar() {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const cameraId = searchParams.get('camera_id')

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

  const handleCameraChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    router.push(pathname + '?' + createQueryString('camera_id', event.target.value))
  }

  return (
    <HStack mb={6} justify="flex-end">
      <chakra.select
        maxW="200px"
        value={cameraId || ''}
        onChange={handleCameraChange}
        p={2}
        borderWidth="1px"
        borderRadius="md"
        bg="white"
      >
        <option value="">Filter by camera</option>
        {CAMERA_IDS.map((id) => (
          <option key={id} value={id}>
            Camera {id}
          </option>
        ))}
      </chakra.select>
    </HStack>
  )
}
