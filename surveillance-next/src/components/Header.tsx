'use client'

import { Box, Container, Flex, Heading, HStack, Link } from '@chakra-ui/react'
import NextLink from 'next/link'

export function Header() {
  return (
    <Box bg="gray.800" color="white" py={4} shadow="md">
      <Container maxW="container.xl">
        <Flex justify="space-between" align="center">
          <Link as={NextLink} href="/" _hover={{ textDecoration: 'none' }}>
            <Heading size="md">🎥 Surveillance Events</Heading>
          </Link>
          
          <HStack gap={6}>
            <Link
              as={NextLink}
              href="/events"
              fontWeight="medium"
              _hover={{ color: 'brand.300' }}
            >
              Events
            </Link>
            <Link
              as={NextLink}
              href="/frames"
              fontWeight="medium"
              _hover={{ color: 'brand.300' }}
            >
              Frames
            </Link>
          </HStack>
        </Flex>
      </Container>
    </Box>
  )
}