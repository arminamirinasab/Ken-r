package ir.kenar.domain.widget

import androidx.annotation.StringRes
import ir.kenar.R

/**
 * Quick, low-friction messages sent from the Love Tap widget.
 *
 * [wireValue] is the locale-independent payload value used by the sync layer.
 * The server stores the encrypted payload as an opaque blob, so Android keeps
 * the user-facing labels here and resolves them per active locale.
 */
enum class LoveTap(
    val wireValue: String,
    @StringRes val labelRes: Int,
    val emoji: String,
    val isTapBack: Boolean = false,
) {
    I_LOVE_YOU("i_love_you", R.string.love_i_love_you, "❤️"),
    I_MISS_YOU("i_miss_you", R.string.love_i_miss_you, "💭"),
    GOOD_NIGHT("good_night", R.string.love_good_night, "🌙"),
    GOOD_MORNING("good_morning", R.string.love_good_morning, "☀️"),
    CAUGHT_IT("caught_it", R.string.love_caught_it, "💛", isTapBack = true);

    companion object {
        /** Primary quick actions shown on the widget surface. */
        val primaryActions: List<LoveTap> = entries.filterNot { it.isTapBack }

        /** Parse a wire value back to a LoveTap, or null if unknown. */
        fun fromWire(value: String?): LoveTap? = entries.firstOrNull { it.wireValue == value }
    }
}
