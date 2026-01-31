# EduPulse

EduPulse is an intelligent educational platform designed to facilitate academic and professional learning. Powered by Google's Gemini AI, it dynamically generates quizzes and study plans based on user-defined topics, analyzes performance to provide personalized feedback, and tracks progress over time.

## Problem Statement

Traditional education systems often struggle to provide personalized attention to every student due to resource constraints. The "one-size-fits-all" approach to testing and study materials can leave students behind or fail to challenge advanced learners. EduPulse addresses this gap by providing an on-demand, adaptive tutor that creates assessments and learning paths tailored to the user's specific grade level, topic of interest, and performance history.

## SDG Alignment

This project directly supports **UN Sustainable Development Goal 4: Quality Education**.
*   **Target 4.1**: By providing free, accessible, and adaptive learning tools, it helps ensure effective learning outcomes.
*   **Target 4.4**: It aids in acquiring relevant skills for employment and entrepreneurship through professional-level assessments.
*   **Target 4.c**: It acts as a force multiplier for educators, reducing the burden of content creation and grading.

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

## Architecture & Design Choices

*   **Monolithic Go Backend**: Chosen for its strong standard library, performance, and ease of deployment. It handles API orchestration, database management, and session logic in a single binary.
*   **Embedded SQLite Database**: Selected for zero-configuration persistence. It allows the application to be self-contained without requiring external database servers, making it ideal for local deployments and educational demos.
*   **Vanilla JavaScript Frontend**: A deliberate choice to avoid complex build chains (npm/webpack) and keep the project lightweight. State management is handled via a simple reactive store pattern in `app.js`.
*   **Stateless AI Interaction**: The system uses a RESTful approach to interact with Gemini, maintaining context via the database rather than keeping long-running socket connections, ensuring scalability.

## AI Usage

EduPulse leverages Google's **Gemini 1.5 Flash** model for three distinct capabilities:
1.  **Content Generation**: Creating structured JSON quizzes and study plans from natural language prompts or uploaded images.
2.  **Semantic Grading**: Evaluating open-ended text responses where simple keyword matching fails, providing human-like feedback.
3.  **Adaptive Recommendations**: Analyzing aggregated user performance data (weaknesses) to suggest targeted remedial topics.

*Safety Guardrails*: System instructions are injected into every prompt to strictly limit the AI to educational contexts and reject inappropriate topics.

## Trade-offs & Limitations

*   **Latency**: Generating high-quality AI content takes time (2-5 seconds). While loading states are implemented, it is slower than retrieving pre-indexed content.
*   **Database Scalability**: SQLite is excellent for this use case but would require migration to PostgreSQL for a high-traffic, multi-server production environment.
*   **AI Hallucinations**: As with all LLMs, there is a non-zero chance of generating factually incorrect questions. Users are advised to verify critical information.
*   **Frontend Complexity**: Managing complex UI states (like the quiz timer and dashboard charts) in Vanilla JS requires verbose DOM manipulation compared to frameworks like React.

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