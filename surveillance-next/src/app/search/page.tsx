'use client'

import { useState, useEffect, useCallback } from 'react'
import {
  Box,
  Container,
  Heading,
  Text,
  SimpleGrid,
  Input,
  Button,
  VStack,
  HStack,
  Spinner,
  Badge,
  Stack,
  NativeSelectRoot,
  NativeSelectField,
} from '@chakra-ui/react'
import { Header } from '@/components/Header'
import { EventCard } from '@/components/EventCard'
import { searchEvents, type Event, type SearchParams } from '@/lib/api'

export default function SearchPage() {
  const [query, setQuery] = useState('')
  const [severity, setSeverity] = useState('')
  const [eventType, setEventType] = useState('')
  const [results, setResults] = useState<Event[]>([])
  const [loading, setLoading] = useState(false)
  const [searched, setSearched] = useState(false)

  const handleSearch = useCallback(async () => {
    setLoading(true)
    setSearched(true)

    const params: SearchParams = {
      limit: 50,
    }

    if (query) params.q = query
    if (severity) params.severity = severity
    if (eventType) params.type = eventType

    try {
      const events = await searchEvents(params)
      setResults(events)
    } catch (error) {
      console.error('Search error:', error)
      setResults([])
    } finally {
      setLoading(false)
    }
  }, [query, severity, eventType])

  const handleClear = () => {
    setQuery('')
    setSeverity('')
    setEventType('')
    setResults([])
    setSearched(false)
  }

  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.key === 'Enter') {
        handleSearch()
      }
    }

    window.addEventListener('keypress', handleKeyPress)
    return () => window.removeEventListener('keypress', handleKeyPress)
  }, [handleSearch])

  return (
    <Box minH="100vh" bg="gray.50">
      <Header />

      <Container maxW="container.xl" py={8}>
        <VStack align="stretch" gap={6}>
          {/* Header */}
          <Box>
            <Heading size="xl" mb={2} color="gray.800">
              🔍 Search Events
            </Heading>
            <Text color="gray.600">
              Search through all detection events using Elasticsearch
            </Text>
          </Box>

          {/* Search Form */}
          <Box bg="white" p={6} borderRadius="lg" shadow="sm" borderWidth="1px">
            <VStack align="stretch" gap={4}>
              {/* Text Search */}
              <Box>
                <Text fontSize="sm" fontWeight="medium" mb={2} color="gray.700">
                  Search Query
                </Text>
                <Input
                  placeholder="Search by title or message..."
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  size="lg"
                />
              </Box>

              {/* Filters */}
              <Stack direction={{ base: 'column', md: 'row' }} gap={4}>
                <Box flex={1}>
                  <Text fontSize="sm" fontWeight="medium" mb={2} color="gray.700">
                    Severity
                  </Text>
                  <NativeSelectRoot size="lg">
                    <NativeSelectField
                      placeholder="All severities"
                      value={severity}
                      onChange={(e) => setSeverity(e.target.value)}
                    >
                      <option value="">All severities</option>
                      <option value="low">Low</option>
                      <option value="medium">Medium</option>
                      <option value="high">High</option>
                      <option value="critical">Critical</option>
                    </NativeSelectField>
                  </NativeSelectRoot>
                </Box>

                <Box flex={1}>
                  <Text fontSize="sm" fontWeight="medium" mb={2} color="gray.700">
                    Event Type
                  </Text>
                  <NativeSelectRoot size="lg">
                    <NativeSelectField
                      placeholder="All types"
                      value={eventType}
                      onChange={(e) => setEventType(e.target.value)}
                    >
                      <option value="">All types</option>
                      <option value="detection">Detection</option>
                      <option value="motion">Motion</option>
                      <option value="system_start">System Start</option>
                      <option value="system_stop">System Stop</option>
                      <option value="camera_online">Camera Online</option>
                      <option value="camera_offline">Camera Offline</option>
                    </NativeSelectField>
                  </NativeSelectRoot>
                </Box>
              </Stack>

              {/* Action Buttons */}
              <HStack gap={3}>
                <Button
                  colorPalette="blue"
                  size="lg"
                  onClick={handleSearch}
                  loading={loading}
                  flex={1}
                >
                  Search
                </Button>
                <Button
                  variant="outline"
                  colorPalette="gray"
                  size="lg"
                  onClick={handleClear}
                >
                  Clear
                </Button>
              </HStack>
            </VStack>
          </Box>

          {/* Results */}
          {loading && (
            <Box textAlign="center" py={12}>
              <Spinner size="xl" color="blue.500" />
              <Text mt={4} color="gray.600">
                Searching...
              </Text>
            </Box>
          )}

          {!loading && searched && (
            <>
              {/* Results Header */}
              <HStack justify="space-between" align="center">
                <Text fontSize="lg" fontWeight="medium" color="gray.700">
                  {results.length} result{results.length !== 1 ? 's' : ''} found
                </Text>
                {(query || severity || eventType) && (
                  <HStack gap={2}>
                    <Text fontSize="sm" color="gray.500">
                      Filters:
                    </Text>
                    {query && (
                      <Badge colorPalette="blue" size="sm">
                        Query: {query}
                      </Badge>
                    )}
                    {severity && (
                      <Badge colorPalette="orange" size="sm">
                        Severity: {severity}
                      </Badge>
                    )}
                    {eventType && (
                      <Badge colorPalette="purple" size="sm">
                        Type: {eventType}
                      </Badge>
                    )}
                  </HStack>
                )}
              </HStack>

              {/* Results Grid */}
              {results.length > 0 ? (
                <SimpleGrid columns={{ base: 1, sm: 2, lg: 3 }} gap={6}>
                  {results.map((event) => (
                    <EventCard key={event.id} event={event} />
                  ))}
                </SimpleGrid>
              ) : (
                <Box
                  bg="white"
                  textAlign="center"
                  py={16}
                  borderRadius="lg"
                  borderWidth="1px"
                >
                  <Text fontSize="3xl" mb={2}>
                    🔍
                  </Text>
                  <Text fontSize="lg" color="gray.600" fontWeight="medium">
                    No events found
                  </Text>
                  <Text fontSize="sm" color="gray.500" mt={1}>
                    Try adjusting your search filters
                  </Text>
                </Box>
              )}
            </>
          )}

          {/* Initial State */}
          {!loading && !searched && (
            <Box
              bg="white"
              textAlign="center"
              py={16}
              borderRadius="lg"
              borderWidth="1px"
            >
              <Text fontSize="3xl" mb={2}>
                🔎
              </Text>
              <Text fontSize="lg" color="gray.600" fontWeight="medium">
                Start searching
              </Text>
              <Text fontSize="sm" color="gray.500" mt={1}>
                Enter your search criteria above and click Search
              </Text>
            </Box>
          )}
        </VStack>
      </Container>
    </Box>
  )
}
