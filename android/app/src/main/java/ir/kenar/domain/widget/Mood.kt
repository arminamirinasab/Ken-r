package ir.kenar.domain.widget

import androidx.annotation.StringRes
import ir.kenar.R

/**
 * Mood states a user can broadcast to their partner's widget.
 * [wireValue] is the stable, locale-independent identifier used in payloads and
 * persistence; [labelRes] is resolved per-locale at render time (fa/en).
 */
enum class Mood(val wireValue: String, @StringRes val labelRes: Int, val emoji: String) {
    HAPPY("happy", R.string.mood_happy, "😊"),
    SAD("sad", R.string.mood_sad, "😢"),
    TIRED("tired", R.string.mood_tired, "😴"),
    LOVING("loving", R.string.mood_loving, "🥰"),
    ANGRY("angry", R.string.mood_angry, "😠");

    companion object {
        /** Parse a wire value back to a Mood, or null if unknown. */
        fun fromWire(value: String?): Mood? = entries.firstOrNull { it.wireValue == value }
    }
}
