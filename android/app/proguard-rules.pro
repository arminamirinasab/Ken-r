# R8 / ProGuard rules. Obfuscation enabled in release builds (security req).

# Keep Glance widget receivers (referenced from the manifest by name).
-keep class ir.kenar.widget.** { *; }

# Hilt generated components.
-keep class dagger.hilt.** { *; }
-keep class * extends androidx.lifecycle.ViewModel { *; }

# Kotlin metadata for reflection-friendly libraries.
-keepattributes RuntimeVisibleAnnotations,AnnotationDefault

# OkHttp / Okio (no warnings for optional platform APIs).
-dontwarn okhttp3.**
-dontwarn okio.**
