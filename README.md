# GalaxyOS

GalaxyOS is an internal monorepo project designed to manage the bots used on the Galaxy One Discord server. Built with Go and utilizing the Disgo library, the aim is to streamline the development, maintenance, and deployment of custom bots (e.g., Kevin and Hue) for specific server functionalities. This README provides documentation and guidance for managing and expanding the project.

Project Purpose:

- Centralized Bot Management: Simplify updates and code sharing between bots under a unified monorepo.
- Customization: Tailored bots to meet the evolving needs of the server, with unique features and commands.
- Efficiency: Reduce redundant code and streamline development with shared modules and common utilities.

This document will guide you through the structure and architecture of the project, as well as key modules and configurations.

## Project structure

```
cmd/ # Contains the main package and a specific configuration for each bot
sdk/ # Contains all shared modules
```
