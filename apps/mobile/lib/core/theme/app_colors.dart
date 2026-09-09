import 'package:flutter/material.dart';

/// Design tokens matching Filvault DESIGN.md
abstract class AppColors {
  // Brand chromatic accent
  static const Color accent = Color(0xFF0D9488); // Filvault Teal
  static const Color accentHover = Color(0xFF0F766E);
  static const Color accentSoft = Color(0xFFCCFBF1);

  // Neutral palette (Light mode)
  static const Color ink = Color(0xFF111827);
  static const Color body = Color(0xFF374151);
  static const Color muted = Color(0xFF6B7280);
  static const Color mutedSoft = Color(0xFF9CA3AF);
  static const Color canvas = Color(0xFFFFFFFF);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color surfaceSoft = Color(0xFFF8FAFC);
  static const Color surfaceCard = Color(0xFFF3F4F6);
  static const Color hairline = Color(0xFFE5E7EB);

  // Status colors
  static const Color danger = Color(0xFFDC2626);
  static const Color dangerSoft = Color(0xFFFEE2E2);
  static const Color warning = Color(0xFFD97706);
  static const Color success = Color(0xFF059669);

  // Dark palette
  static const Color darkCanvas = Color(0xFF121212);
  static const Color darkSurface = Color(0xFF18191A);
  static const Color darkSurfaceSoft = Color(0xFF242526);
  static const Color darkSurfaceCard = Color(0xFF3A3B3C);
  static const Color darkHairline = Color(0xFF2E2F30);
  static const Color darkInk = Color(0xFFE4E6EB);
  static const Color darkMuted = Color(0xFFB0B3B8);
}
