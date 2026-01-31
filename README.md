# EduPulse AI (Go Version)

EduPulse AI is an intelligent educational platform designed to facilitate academic and professional learning. Powered by Google's Gemini AI, it dynamically generates quizzes and study plans based on user-defined topics, analyzes performance to provide personalized feedback, and tracks progress over time.

## Features

*   **Authentication & User Management**:
    *   Secure Sign Up and Login.
    *   **Guest Mode** for quick access without an account.
    *   Profile management (update details, change password).
*   **Personalized Dashboard**:
    *   Overview of total quizzes taken and average score.
    *   Visual charts for recent activity.
    *   **AI Recommendations** based on identified weaknesses (cached for efficiency).
*   **AI-Powered Content Generation**:
    *   **Quizzes**: Create assessments on any topic with adjustable difficulty (Elementary to Professional).
    *   **Study Plans**: Generate structured learning modules with time estimates and activities.
*   **Flexible Assessment Formats**:
    *   Multiple Choice (MCQ), Open-Ended, or Mixed question types.
    *   **Image Analysis**: Upload study notes or diagrams to generate relevant questions.
*   **Adaptive Modes**:
    *   *Per Question*: Timed strictly per question.
    *   *Full Session*: Manage time across the entire quiz.
*   **Comprehensive Analysis**:
    *   Instant grading and feedback.
    *   Identification of strengths and weaknesses.
    *   Performance tracking stored in a local SQLite database.

## Tech Stack

*   **Backend**: Go (Golang)
*   **Database**: SQLite (`edupulse.db`)
*   **Frontend**: Vanilla JavaScript, HTML5, Tailwind CSS
*   **AI Engine**: Google Gemini API (`gemini-3-flash-preview`)
*   **Security**: Bcrypt password hashing

## Prerequisites

*   Go (1.21 or later recommended)
*   GCC (required for `go-sqlite3`)
*   A Google Cloud Project with the Gemini API enabled
*   An API Key for Gemini

## Setup & Installation

1.  **Navigate to the project directory**:
    ```bash
    cd "/home/tomlee/Desktop/dev/Go version"
    ```

2.  **Install Dependencies**:
    Ensure your `go.mod` is set up, then run:
    ```bash
    go mod tidy
    ```

3.  **Environment Configuration**:
    Create a `.env` file in the root directory and add your Gemini API key:
    ```env
    GEMINI_API_KEY=your_api_key_here
    PORT=8080
    ```

4.  **Run the Application**:
    ```bash
    go run main.go
    ```
    *Note: The application will automatically initialize the SQLite database (`edupulse.db`) on the first run.*

5.  **Access the App**:
    Open your browser and navigate to `http://localhost:8080`.

## Usage Guide

1.  **Authentication**: Sign up for an account, log in, or continue as a Guest.
2.  **Dashboard**: View your stats and AI-curated recommendations.
3.  **Create Content**:
    *   Click "Start New Session".
    *   Choose **Take a Quiz** or **Generate Study Plan**.
    *   Enter a topic (e.g., "Photosynthesis") or upload an image.
    *   Configure settings (Question type, Timing).
4.  **Assessment/Learning**: Complete the quiz or review the study plan.
5.  **Analysis**: Receive instant feedback and return to the dashboard to see updated stats.

## Project Structure

*   `main.go`: The core Go server handling API routes, Gemini integration, database logic, and session management.
*   `static/`: Contains frontend assets.
    *   `js/app.js`: Client-side logic for state management, UI rendering, and API calls.
    *   `index.html`: Main entry point.
*   `edupulse.db`: Local SQLite database (created at runtime).