
# EduPulse AI - Intelligent Learning Platform

EduPulse AI is a cutting-edge educational platform that leverages Google's Gemini 3 Flash model to create personalized, context-aware learning experiences.

## 🚀 Features

- **AI Quiz Generation**: Dynamically creates quizzes based on student grade, interests, and goals.
- **Multimodal Support**: Upload study notes (images) to ground the AI in specific course material.
- **Pedagogical Analysis**: Deep-dive performance metrics with AI-generated study recommendations.
- **Authentication**: Secure login/logout flow with session persistence.
- **Real-time Metrics**: Tracks per-question time consumption and revisit patterns.

## 🛠 Tech Stack

- **Frontend**: React (TypeScript), Tailwind CSS, Recharts.
- **Backend**: Go (Conceptual `main.go` provided for production scaling).
- **AI**: Google Gemini API (@google/genai).
- **Database**: Simulated via LocalStorage (Frontend) / Planned PostgreSQL (Backend).

## 🏁 Getting Started

### Prerequisites
- Node.js & npm (for frontend)
- Go 1.21+ (for backend)
- Gemini API Key

### Installation
1. Clone the repository.
2. Set your environment variable: `export API_KEY='your_gemini_key'`.
3. For the frontend: The app is configured to run as an ES module via `index.html`.
4. For the backend:
   ```bash
   go run main.go
   ```

## 📈 Roadmap
- [ ] Integration with real PostgreSQL database.
- [ ] Collaborative study rooms using Live API.
- [ ] PDF parsing for study material grounding.
