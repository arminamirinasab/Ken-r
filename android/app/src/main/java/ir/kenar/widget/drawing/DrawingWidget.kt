package ir.kenar.widget.drawing

import android.content.Context
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.glance.GlanceId
import androidx.glance.GlanceModifier
import androidx.glance.GlanceTheme
import androidx.glance.Image
import androidx.glance.ImageProvider
import androidx.glance.LocalContext
import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.provideContent
import androidx.glance.background
import androidx.glance.currentState
import androidx.glance.layout.Alignment
import androidx.glance.layout.Column
import androidx.glance.layout.ContentScale
import androidx.glance.layout.fillMaxSize
import androidx.glance.layout.fillMaxWidth
import androidx.glance.layout.padding
import androidx.glance.state.GlanceStateDefinition
import androidx.glance.state.PreferencesGlanceStateDefinition
import androidx.glance.text.Text
import androidx.glance.text.TextStyle
import androidx.glance.unit.ColorProvider
import ir.kenar.R
import ir.kenar.domain.widget.SharedDrawing

/**
 * Passive home-screen renderer for the partner's latest shared drawing.
 *
 * The sync layer writes [PARTNER_DRAWING_KEY] after decrypting/fetching state
 * from the generic `drawing` widget endpoint. This widget only renders that
 * local state and never performs network work.
 */
class DrawingWidget : GlanceAppWidget() {

    override val stateDefinition: GlanceStateDefinition<*> = PreferencesGlanceStateDefinition

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        provideContent { Content() }
    }

    @Composable
    private fun Content() {
        val context = LocalContext.current
        val drawing = SharedDrawing.fromWire(currentState(PARTNER_DRAWING_KEY))
        val updatedAt = currentState(PARTNER_DRAWING_UPDATED_AT_KEY)

        Column(
            modifier = GlanceModifier
                .fillMaxSize()
                .background(GlanceTheme.colors.widgetBackground)
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                text = context.getString(R.string.drawing_title),
                style = TextStyle(color = onSurface()),
            )

            if (drawing != null && !drawing.isEmpty()) {
                Image(
                    provider = ImageProvider(DrawingPreviewRenderer.render(drawing)),
                    contentDescription = context.getString(R.string.drawing_content_description),
                    modifier = GlanceModifier.fillMaxWidth(),
                    contentScale = ContentScale.Fit,
                )
                if (updatedAt != null) {
                    Text(
                        text = context.getString(R.string.drawing_updated),
                        style = TextStyle(color = onSurfaceVariant()),
                    )
                }
            } else {
                Text(
                    text = context.getString(R.string.drawing_waiting),
                    style = TextStyle(color = onSurfaceVariant()),
                )
            }
        }
    }

    private fun onSurface(): ColorProvider = ColorProvider(Color(0xFF2B1A1F))

    private fun onSurfaceVariant(): ColorProvider = ColorProvider(Color(0xFF6D5961))

    companion object {
        /** Partner's latest drawing wire payload. Written by sync. */
        val PARTNER_DRAWING_KEY = stringPreferencesKey("partner_drawing")
        val PARTNER_DRAWING_UPDATED_AT_KEY = stringPreferencesKey("partner_drawing_updated_at")

        /** Latest local drawing waiting for encryption/upload by sync. */
        val OUTGOING_DRAWING_KEY = stringPreferencesKey("outgoing_drawing")
        val OUTGOING_DRAWING_CREATED_AT_KEY = stringPreferencesKey("outgoing_drawing_created_at")
    }
}
