# Changelog

## [Unreleased]

### Added
- Enhanced `Place` struct in `internal/itinerary` to include `ImageURL`, `Description`, `Country`, and `Rating`.
- Updated `GenerateItinerary` to populate new fields from Pinecone metadata.
- Added `Map` component to the Itinerary page (`frontend/app/itinerary/page.tsx`).
- Added `ResultCard` component usage to the Itinerary page for better list display.

### Changed
- Improved itinerary generation algorithm in `internal/itinerary/itinerary.go` to split places into days using chunks instead of round-robin, preserving spatial continuity.
- Updated Itinerary page frontend to display a map and center it on the first location.
