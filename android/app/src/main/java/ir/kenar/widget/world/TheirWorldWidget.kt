package ir.kenar.widget.world

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
import androidx.glance.layout.Row
import androidx.glance.layout.fillMaxSize
import androidx.glance.layout.padding
import androidx.glance.state.GlanceStateDefinition
import androidx.glance.state.PreferencesGlanceStateDefinition
import androidx.glance.text.Text
import androidx.glance.text.TextStyle
import androidx.glance.unit.ColorProvider
import ir.kenar.R
import ir.kenar.domain.widget.TheirWorld

/**
 * Passive renderer for the partner's local time, weather, and day/night state.
 *
 * Weather fetches and timezone conversion happen in the app sync layer. This
 * widget only displays the latest privacy-safe snapshot.
 */
class TheirWorldWidget : GlanceAppWidget() {

    override val stateDefinition: GlanceStateDefinition<*> = PreferencesGlanceStateDefinition

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        provideContent { Content() }
    }

    @Composable
    private fun Content() {
        val context = LocalContext.current
        val world = TheirWorld.fromWire(currentState(PARTNER_WORLD_KEY))

        Column(
            modifier = GlanceModifier
                .fillMaxSize()
                .background(GlanceTheme.colors.widgetBackground)
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                text = context.getString(R.string.world_title),
                style = TextStyle(color = onSurface()),
            )

            if (world != null) {
                Text(
                    text = world.locationLabel,
                    style = TextStyle(color = onSurfaceVariant()),
                )
                Text(
                    text = context.getString(R.string.world_local_time, world.localTimeLabel),
                    style = TextStyle(color = onAccent()),
                )
                Row(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        text = world.weather.glyph,
                        style = TextStyle(color = onSurface()),
                    )
                    Text(
                        text = world.weatherLine(context),
                        style = TextStyle(color = onSurface()),
                    )
                }
                Text(
                    text = context.getString(world.dayNight.labelRes),
                    style = TextStyle(color = onSurfaceVariant()),
                )
            } else {
                Text(
                    text = context.getString(R.string.world_waiting),
                    style = TextStyle(color = onSurfaceVariant()),
                )
            }
        }
    }

    private fun TheirWorld.weatherLine(context: Context): String {
        val weatherLabel = context.getString(weather.labelRes)
        return if (temperatureCelsius != null) {
            context.getString(R.string.world_weather_with_temp, weatherLabel, temperatureCelsius)
        } else {
            weatherLabel
        }
    }

    private fun onSurface(): ColorProvider = ColorProvider(Color(0xFF2B1A1F))

    private fun onSurfaceVariant(): ColorProvider = ColorProvider(Color(0xFF6D5961))

    private fun onAccent(): ColorProvider = ColorProvider(Color(0xFFE04F7A))

    companion object {
        /** Partner's latest world snapshot. Written by sync. */
        val PARTNER_WORLD_KEY = stringPreferencesKey("partner_world")

        /** Latest local world snapshot waiting for encryption/upload by sync. */
        val OUTGOING_WORLD_KEY = stringPreferencesKey("outgoing_world")
        val OUTGOING_WORLD_CREATED_AT_KEY = stringPreferencesKey("outgoing_world_created_at")
    }
}
