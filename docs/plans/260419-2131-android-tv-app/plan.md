# Plan: Velox Android TV App
Created: 2026-04-19 21:31
Status: 🟡 In Progress

## Overview
Xây dựng phiên bản Android TV cho Velox Media Server. Tính năng được focus hoàn toàn vào mục đích tiêu thụ nội dung (Consumption): Tìm kiếm, Duyệt danh mục, và Xem phim.
LOẠI BỎ hoàn toàn các tính năng Admin/Server Settings theo yêu cầu. Giao diện được tối ưu 100% cho điều khiển D-Pad Remote sử dụng Jetpack Compose for TV.

## Tech Stack
- Frontend: Jetpack Compose for TV (androidx.tv.material3, androidx.tv.foundation)
- Backend: Sẵn có (Tái sử dụng Velox API)
- Storage & Data: Tái sử dụng Layer Data & Domain của App Mobile hiện tại.
- Video Player: ExoPlayer (có sẵn, custom thêm TV OSD Controller)

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | Setup Environment & Manifest | ✅ Complete | 100% |
| 02 | Login & Home Dashboard | ✅ Complete | 100% |
| 03 | Browse, Search & Details | ✅ Complete | 100% |
| 04 | Player OSD & Actions | ✅ Complete | 100% |

## Quick Commands
- Start Phase 1: `/code phase-01`
- Check progress: `/next`
- Save context: `/save-brain`
