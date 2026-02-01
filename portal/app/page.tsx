import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export default function Home() {
  const models = [
    { name: 'GPT-4o', provider: 'OpenAI', category: 'Chat', price: '$0.005 / 1K tokens', context: '128K', capabilities: ['Text', 'Vision'] },
    { name: 'Claude 3 Sonnet', provider: 'Anthropic', category: 'Chat', price: '$0.003 / 1K tokens', context: '200K', capabilities: ['Text', 'Vision'] },
    { name: 'Gemini 1.5 Pro', provider: 'Google', category: 'Chat', price: '$0.00125 / 1K tokens', context: '1M', capabilities: ['Text', 'Vision', 'Audio'] },
    { name: 'Llama 3 70B', provider: 'Meta', category: 'Chat', price: '$0.0009 / 1K tokens', context: '8K', capabilities: ['Text'] },
    { name: 'Command R', provider: 'Cohere', category: 'Chat', price: '$0.0005 / 1K tokens', context: '128K', capabilities: ['Text'] },
    { name: 'GPT-3.5 Turbo', provider: 'OpenAI', category: 'Chat', price: '$0.0005 / 1K tokens', context: '16K', capabilities: ['Text'] },
  ]

  return (
    <>
      {/* Hero Section */}
      <section className="py-20 px-4 text-center">
        <h1 className="text-5xl font-bold tracking-tight mb-6">
          Unified Access to <span className="text-primary">300+</span> AI Models
        </h1>
        <p className="text-xl text-muted-foreground max-w-3xl mx-auto mb-10">
          MassRouter SaaS provides a single API for all major AI models. Streamline your development with unified pricing, billing, and monitoring.
        </p>
        <div className="flex gap-4 justify-center">
          <Button size="lg">Get Started</Button>
          <Button size="lg" variant="outline">View Documentation</Button>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-16 px-4 bg-muted/30">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-3xl font-bold text-center mb-12">Why Choose MassRouter SaaS?</h2>
          <div className="grid md:grid-cols-3 gap-8">
            <Card>
              <CardHeader>
                <CardTitle>Unified API</CardTitle>
                <CardDescription>One interface for all models</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Access 60+ providers with a single API key. No need to manage multiple accounts or integrations.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Smart Routing</CardTitle>
                <CardDescription>Automatic failover and load balancing</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Intelligent routing ensures high availability and optimal performance across all providers.
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>Usage Analytics</CardTitle>
                <CardDescription>Detailed insights and reporting</CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Monitor usage, costs, and performance across all models with comprehensive dashboards.
                </p>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>

      {/* Models Section */}
      <section className="py-16 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="flex justify-between items-center mb-8">
            <h2 className="text-3xl font-bold">Popular Models</h2>
            <Button variant="outline">View All Models</Button>
          </div>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {models.map((model, index) => (
              <Card key={index} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="flex justify-between items-start">
                    <div>
                      <CardTitle>{model.name}</CardTitle>
                      <CardDescription>{model.provider}</CardDescription>
                    </div>
                    <Badge variant="secondary">{model.category}</Badge>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div>
                      <p className="text-sm text-muted-foreground">Price</p>
                      <p className="text-2xl font-bold">{model.price}</p>
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground">Context Length</p>
                      <p className="font-medium">{model.context}</p>
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground mb-2">Capabilities</p>
                      <div className="flex flex-wrap gap-2">
                        {model.capabilities.map((cap, i) => (
                          <Badge key={i} variant="outline">{cap}</Badge>
                        ))}
                      </div>
                    </div>
                    <Button className="w-full">Try Now</Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 px-4 bg-primary text-primary-foreground">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-4xl font-bold mb-6">Ready to Get Started?</h2>
          <p className="text-xl mb-10 opacity-90">
            Join thousands of developers building with the most comprehensive AI platform.
          </p>
          <div className="flex gap-4 justify-center">
            <Button size="lg" variant="secondary">Sign Up Free</Button>
            <Button size="lg" variant="outline" className="bg-transparent border-white text-white hover:bg-white/10">
              Contact Sales
            </Button>
          </div>
        </div>
      </section>

    </>
  )
}