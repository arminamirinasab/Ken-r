package ir.kenar.widget.mood

import androidx.compose.runtime.Composable
import androidx.glance.GlanceModifier
import androidx.glance.GlanceTheme
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.datastore.preferences.core.stringPreferencesKey
import android.content.Context
import androidx.glance.GlanceId
import androidx.glance.LocalContext
import ir.kenar.R
import ir.kenar.domain.widget.Mood

/**
 * Passive, stateless home-screen widget showing the PARTNER's latest mood.
 *
 * The widget never fetches or computes — it only renders the value written into
 * its Glance state by the sync layer (WebSocket message or a Pushe wake →
 * fetch). This keeps it battery-friendly and within widget IPC/bitmap limits.
 */
class MoodWidget : GlanceAppWidget() {

    override val stateDefinition: GlanceStateDefinition<*> = PreferencesGlanceStateDefinition

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        provideContent { Content() }
    }

    @Composable
    private fun Content() {
        val context = LocalContext.current
        val wire = currentState(PARTNER_MOOD_KEY)
        val mood = Mood.fromWire(wire)

        Column(
            modifier = GlanceModifier
                .fillMaxSize()
                .background(GlanceTheme.colors.widgetBackground)
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            if (mood != null) {
                Text(text = mood.emoji, style = TextStyle(color = onSurface()))
                Text(
                    text = context.getString(mood.labelRes),
                    style = TextStyle(color = onSurface()),
                )
            } else {
                // Beautiful empty state until the partner shares a mood.
                Text(
                    text = context.getString(R.string.mood_waiting),
                    style = TextStyle(color = onSurface()),
                )
            }
        }
    }

    private fun onSurface(): ColorProvider = ColorProvider(Color(0xFF2B1A1F))

    companion object {
        /** Glance state key holding the partner's latest mood wire value. */
        val PARTNER_MOOD_KEY = stringPreferencesKey("partner_mood")
    }
}
