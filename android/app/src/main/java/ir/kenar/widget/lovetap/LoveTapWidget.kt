package ir.kenar.widget.lovetap

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.glance.Button
import androidx.glance.GlanceId
import androidx.glance.GlanceModifier
import androidx.glance.GlanceTheme
import androidx.glance.LocalContext
import androidx.glance.action.ActionParameters
import androidx.glance.action.actionParametersOf
import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.action.actionRunCallback
import androidx.glance.appwidget.provideContent
import androidx.glance.background
import androidx.glance.currentState
import androidx.glance.layout.Alignment
import androidx.glance.layout.Column
import androidx.glance.layout.Row
import androidx.glance.layout.fillMaxSize
import androidx.glance.layout.fillMaxWidth
import androidx.glance.layout.padding
import androidx.glance.state.GlanceStateDefinition
import androidx.glance.state.PreferencesGlanceStateDefinition
import androidx.glance.text.Text
import androidx.glance.text.TextStyle
import androidx.glance.unit.ColorProvider
import ir.kenar.R
import ir.kenar.domain.widget.LoveTap

/**
 * Home-screen widget for fast affectionate taps and the tap-back loop.
 *
 * Received taps are written by the sync layer into [PARTNER_TAP_KEY]. User
 * actions are queued in [OUTGOING_TAP_KEY] for the sync layer to encrypt and
 * upload through the generic `love_tap` widget endpoint.
 */
class LoveTapWidget : GlanceAppWidget() {

    override val stateDefinition: GlanceStateDefinition<*> = PreferencesGlanceStateDefinition

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        provideContent { Content() }
    }

    @Composable
    private fun Content() {
        val context = LocalContext.current
        val partnerTap = LoveTap.fromWire(currentState(PARTNER_TAP_KEY))
        val lastSentTap = LoveTap.fromWire(currentState(LAST_SENT_TAP_KEY))

        Column(
            modifier = GlanceModifier
                .fillMaxSize()
                .background(GlanceTheme.colors.widgetBackground)
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                text = context.getString(R.string.love_title),
                style = TextStyle(color = onSurface()),
            )

            if (partnerTap != null) {
                Text(
                    text = "${partnerTap.emoji} ${context.getString(partnerTap.labelRes)}",
                    style = TextStyle(color = onSurface()),
                )
                LoveTapButton(tap = LoveTap.CAUGHT_IT, label = context.getString(R.string.love_action_respond))
            } else {
                Text(
                    text = context.getString(R.string.love_waiting),
                    style = TextStyle(color = onSurfaceVariant()),
                )
            }

            Row(modifier = GlanceModifier.fillMaxWidth(), horizontalAlignment = Alignment.CenterHorizontally) {
                LoveTap.primaryActions.take(2).forEach { tap ->
                    LoveTapButton(tap = tap, label = "${tap.emoji} ${context.getString(tap.labelRes)}")
                }
            }
            Row(modifier = GlanceModifier.fillMaxWidth(), horizontalAlignment = Alignment.CenterHorizontally) {
                LoveTap.primaryActions.drop(2).forEach { tap ->
                    LoveTapButton(tap = tap, label = "${tap.emoji} ${context.getString(tap.labelRes)}")
                }
            }

            if (lastSentTap != null) {
                Text(
                    text = context.getString(R.string.love_last_sent, context.getString(lastSentTap.labelRes)),
                    style = TextStyle(color = onSurfaceVariant()),
                )
            }
        }
    }

    @Composable
    private fun LoveTapButton(tap: LoveTap, label: String) {
        Button(
            text = label,
            onClick = actionRunCallback<LoveTapActionCallback>(
                parameters = actionParametersOf(TAP_ACTION_PARAM to tap.wireValue),
            ),
        )
    }

    private fun onSurface(): ColorProvider = ColorProvider(Color(0xFF2B1A1F))

    private fun onSurfaceVariant(): ColorProvider = ColorProvider(Color(0xFF6D5961))

    companion object {
        val TAP_ACTION_PARAM = ActionParameters.Key<String>("love_tap_action")

        /** Partner's latest received Love Tap wire value. Written by sync. */
        val PARTNER_TAP_KEY = stringPreferencesKey("partner_love_tap")

        /** Latest user tap waiting for encryption/upload by sync. */
        val OUTGOING_TAP_KEY = stringPreferencesKey("outgoing_love_tap")
        val OUTGOING_TAP_CREATED_AT_KEY = stringPreferencesKey("outgoing_love_tap_created_at")

        /** Immediate local feedback after the user taps a quick action. */
        val LAST_SENT_TAP_KEY = stringPreferencesKey("last_sent_love_tap")
    }
}
