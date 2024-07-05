require 'opentracing'
require 'jaeger/client'
require 'sinatra/base'
require 'logger'

# Initialize Jaeger tracer
tracer = Jaeger::Client.build(
  service_name: 'ruby-example',
  reporter: Jaeger::Reporters::LoggingReporter.new,
  sampler: Jaeger::Samplers::Const.new(true)
)
OpenTracing.global_tracer = tracer

# Initialize logger
logger = Logger.new(STDOUT)
logger.formatter = proc do |severity, datetime, progname, msg|
  span = OpenTracing.active_span
  trace_id = span ? span.context.trace_id : 'N/A'
  span_id = span ? span.context.span_id : 'N/A'
  "[#{datetime}] #{severity} [trace_id=#{trace_id} span_id=#{span_id}]: #{msg}\n"
end


# Sinatra app
class MyApp < Sinatra::Base
    set :bind, '0.0.0.0'
    set :port, 4567

  get '/' do
    OpenTracing.start_active_span('handle_request') do |scope|
      logger.info("Handling request")

      # Simulate some work
      sleep(0.1)

      # Make a downstream request
      OpenTracing.start_active_span('downstream_request') do |inner_scope|
        logger.info("Making downstream request")
        # Simulate downstream request
        sleep(0.2)
        logger.info("Downstream request completed")
      end

      logger.info("Request handled successfully")
      "Hello from Ruby!"
    end
  end
end

# Don't forget to allow time for the reporter to send traces
at_exit do
  sleep 1
end

# Run the Sinatra app
MyApp.run!

