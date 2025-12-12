'use client'

import { Box, Button, Container, Heading, Text, VStack } from '@chakra-ui/react'
import { useRouter } from 'next/navigation'
import { Header } from './Header'

interface NotFoundProps {
  title: string
  message: string
}

export function NotFound({ title, message }: NotFoundProps) {
  const router = useRouter()

  return (
    <Box minH="100vh">
      <Header />
      <Container maxW="container.xl" py={8} centerContent>
        <VStack gap={4} textAlign="center">
          <Heading>{title}</Heading>
          <Text>{message}</Text>
          <Button onClick={() => router.back()} colorScheme="blue">
            Go Back
          </Button>
        </VStack>
      </Container>
    </Box>
  )
}
