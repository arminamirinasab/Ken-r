package ir.kenar.core.locale

import androidx.appcompat.app.AppCompatDelegate
import androidx.core.os.LocaleListCompat

/**
 * In-app language switch. Persian is the product default; the app also respects
 * the system locale until the user explicitly chooses one here.
 *
 * Uses AndroidX per-app locales (autoStoreLocales enabled in the manifest), so
 * the choice persists across launches without a custom store.
 */
object LocaleManager {

    /** Supported app languages, primary first. */
    enum class Language(val tag: String) {
        PERSIAN("fa"),
        ENGLISH("en"),
    }

    /** Apply [language] immediately and persist it. */
    fun set(language: Language) {
        AppCompatDelegate.setApplicationLocales(
            LocaleListCompat.forLanguageTags(language.tag),
        )
    }

    /** Follow the system locale again (clears the explicit app choice). */
    fun followSystem() {
        AppCompatDelegate.setApplicationLocales(LocaleListCompat.getEmptyLocaleList())
    }

    /** The currently effective app language tag (e.g. "fa"), or null if system. */
    fun current(): String? =
        AppCompatDelegate.getApplicationLocales().toLanguageTags().ifEmpty { null }
}
