'use client'

import { Box, Code, Heading, Text } from '@chakra-ui/react'

interface MetadataViewerProps {
  metadata: unknown
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

export function MetadataViewer({ metadata }: MetadataViewerProps) {
  if (!isObject(metadata) || Object.keys(metadata).length === 0) {
    return (
      <Box>
        <Heading size="md" mb={3}>
          Metadata
        </Heading>
        <Text color="gray.500" fontSize="sm">
          No metadata available for this event.
        </Text>
      </Box>
    )
  }

  return (
    <Box>
      <Heading size="md" mb={3}>
        Metadata
      </Heading>
      <Code
        p={4}
        borderRadius="md"
        display="block"
        whiteSpace="pre-wrap"
        wordBreak="break-all"
        fontSize="sm"
        bg="gray.50"
      >
        {JSON.stringify(metadata, null, 2)}
      </Code>
    </Box>
  )
}
