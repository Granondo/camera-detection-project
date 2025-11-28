import { Box, Container, Heading, Text, SimpleGrid, Flex } from '@chakra-ui/react'
import Link from 'next/link'
import { Header } from '@/components/Header'

export default function HomePage() {
  return (
    <Box minH="100vh">
      <Header />
      
      <Container maxW="container.xl" py={10}>
        <Box textAlign="center" mb={12}>
          <Heading size="2xl" mb={4} color="gray.800">
            Video Surveillance Events
          </Heading>
          <Text fontSize="lg" color="gray.600" maxW="600px" mx="auto">
            Public viewer for detection events from video surveillance system.
            Browse events and frames with detected objects.
          </Text>
        </Box>

        <SimpleGrid columns={{ base: 1, md: 2 }} gap={6} maxW="800px" mx="auto">
          <Link href="/events">
            <Box
              bg="white"
              p={8}
              borderRadius="xl"
              shadow="sm"
              transition="all 0.2s"
              _hover={{ shadow: 'lg', transform: 'translateY(-4px)' }}
            >
              <Flex direction="column" align="center" textAlign="center">
                <Text fontSize="4xl" mb={4}>📋</Text>
                <Heading size="md" mb={2} color="gray.800">
                  Events
                </Heading>
                <Text color="gray.600">
                  Browse detection events grouped by type and time
                </Text>
              </Flex>
            </Box>
          </Link>

          <Link href="/frames">
            <Box
              bg="white"
              p={8}
              borderRadius="xl"
              shadow="sm"
              transition="all 0.2s"
              _hover={{ shadow: 'lg', transform: 'translateY(-4px)' }}
            >
              <Flex direction="column" align="center" textAlign="center">
                <Text fontSize="4xl" mb={4}>🖼️</Text>
                <Heading size="md" mb={2} color="gray.800">
                  Frames
                </Heading>
                <Text color="gray.600">
                  Gallery of frames with detected objects
                </Text>
              </Flex>
            </Box>
          </Link>
        </SimpleGrid>
      </Container>
    </Box>
  )
}