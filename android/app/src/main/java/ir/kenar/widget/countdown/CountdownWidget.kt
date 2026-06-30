package ir.kenar.widget.countdown

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.glance.GlanceId
import androidx.glance.GlanceModifier
import androidx.glance.GlanceTheme
import androidx.glance.LocalContext
import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.provideContent
import androidx.glance.background
import androidx.glance.currentState
import androidx.glance.layout.Alignment
import androidx.glance.layout.Column
import androidx.glance.layout.fillMaxSize
import androidx.glance.layout.padding
import androidx.glance.state.GlanceStateDefinition
import androidx.glance.state.PreferencesGlanceStateDefinition
import androidx.glance.text.Text
import androidx.glance.text.TextStyle
import androidx.glance.unit.ColorProvider
import ir.kenar.R
import ir.kenar.domain.widget.Countdown
import ir.kenar.domain.widget.CountdownState
import kotlin.math.abs

/**
 * Passive renderer for a shared date countdown.
 *
 * Sync writes a precomputed [Countdown] snapshot, including localized calendar
 * decisions such as Jalali/Gregorian handling. The widget only renders it.
 */
class CountdownWidget : GlanceAppWidget() {

    override val stateDefinition: GlanceStateDefinition<*> = PreferencesGlanceStateDefinition

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        provideContent { Content() }
    }

    @Composable
    private fun Content() {
        val context = LocalContext.current
        val countdown = Countdown.fromWire(currentState(SHARED_COUNTDOWN_KEY))

        Column(
            modifier = GlanceModifier
                .fillMaxSize()
                .background(GlanceTheme.colors.widgetBackground)
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            if (countdown != null) {
                Text(
                    text = countdown.title,
                    style = TextStyle(color = onSurface()),
                )
                Text(
                    text = countdown.remainingText(context),
                    style = TextStyle(color = onAccent()),
                )
                if (countdown.targetLabel.isNotBlank()) {
                    Text(
                        text = countdown.targetLabel,
                        style = TextStyle(color = onSurfaceVariant()),
                    )
                }
            } else {
                Text(
                    text = context.getString(R.string.countdown_waiting),
                    style = TextStyle(color = onSurfaceVariant()),
                )
            }
        }
    }

    private fun Countdown.remainingText(context: Context): String =
        when (state) {
            CountdownState.UPCOMING -> context.resources.getQuantityString(
                R.plurals.countdown_days_remaining,
                remainingDays,
                remainingDays,
            )
            CountdownState.TODAY -> context.getString(R.string.countdown_today)
            CountdownState.PASSED -> {
                val elapsed = abs(remainingDays)
                context.resources.getQuantityString(R.plurals.countdown_days_passed, elapsed, elapsed)
            }
        }

    private fun onSurface(): ColorProvider = ColorProvider(Color(0xFF2B1A1F))

    private fun onSurfaceVariant(): ColorProvider = ColorProvider(Color(0xFF6D5961))

    private fun onAccent(): ColorProvider = ColorProvider(Color(0xFFE04F7A))

    companion object {
        /** Shared countdown snapshot. Written by sync. */
        val SHARED_COUNTDOWN_KEY = stringPreferencesKey("shared_countdown")

        /** Latest local countdown waiting for encryption/upload by sync. */
        val OUTGOING_COUNTDOWN_KEY = stringPreferencesKey("outgoing_countdown")
        val OUTGOING_COUNTDOWN_CREATED_AT_KEY = stringPreferencesKey("outgoing_countdown_created_at")
    }
}
