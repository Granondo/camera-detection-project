'use client'

import { Box, Container, Flex, HStack, Text } from '@chakra-ui/react'
import NextLink from 'next/link'

export function Header() {
  return (
    <Box bg="gray.800" color="white" shadow="md">
      <Container maxW="container.xl" py={4}>
        <Flex justify="space-between" align="center">
          <NextLink href="/" style={{ textDecoration: 'none', color: 'inherit' }}>
            <Text fontSize="xl" fontWeight="bold">
              🎥 Surveillance
            </Text>
          </NextLink>
          
          <HStack gap={6}>
            <NextLink href="/events" style={{ color: 'inherit' }}>
              Events
            </NextLink>
            <NextLink href="/frames" style={{ color: 'inherit' }}>
              Frames
            </NextLink>
            <NextLink href="/search" style={{ color: 'inherit' }}>
              🔍 Search
            </NextLink>
          </HStack>
        </Flex>
      </Container>
    </Box>
  )
}