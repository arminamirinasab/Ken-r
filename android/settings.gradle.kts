pluginManagement {
    repositories {
        // Iran-friendly: avoid Google Play Services, but Maven repos for the
        // Android Gradle Plugin / Compose are still required to build. Mirror
        // these to a local/self-hosted Nexus if the public ones are blocked.
        google {
            content {
                includeGroupByRegex("com\\.android.*")
                includeGroupByRegex("com\\.google.*")
                includeGroupByRegex("androidx.*")
            }
        }
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "Kenar"
include(":app")
