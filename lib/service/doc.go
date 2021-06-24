// Package service holds the business logic allowing the manipulation of assets.
// The main entry point is the ServiceProvider, from which the services are accessible.
// Services can rely on each other when multiple assets are involved in the same transaction.
// Storage is isolated in the persistence layer and services should only deal with their dedicated asset's persistence.
package service
