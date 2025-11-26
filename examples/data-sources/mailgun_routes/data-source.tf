# List all routes
data "mailgun_routes" "all" {}

output "route_expressions" {
  value = [for r in data.mailgun_routes.all.routes : r.expression]
}

# List routes with a limit
data "mailgun_routes" "limited" {
  limit = 10
}
