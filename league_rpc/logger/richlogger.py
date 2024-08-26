"""
This module provides the RichLogger class for logging with colorful output using the Rich library.
"""

import logging
from typing import Any, Dict, List, Optional

from rich.console import Console
from rich.logging import RichHandler
from rich.progress import BarColumn, Progress, TaskID, TextColumn
from rich.table import Table


class RichLogger:
    """A logger that utilizes the Rich library for colorful, formatted output."""

    def __init__(self, name: str = "LeagueRPC", show_debugs: bool = False) -> None:
        # Create a Console instance for rich output
        self.console: Console = Console()

        # Custom RichHandler without default time and file info
        rich_handler = RichHandler(
            console=self.console,
            rich_tracebacks=True,
            show_time=False,
            show_level=False,
            show_path=False,
        )

        # Setup logging with rich
        logging.basicConfig(
            level=logging.INFO,
            format="%(message)s",
            handlers=[rich_handler],
        )

        # Initialize the logger
        self.logger: logging.Logger = logging.getLogger(name)
        self.progress: Optional[Progress] = None
        self.task: Optional[TaskID] = None
        self.show_debugs: bool = show_debugs

    def debug(
        self,
        message: str,
        *args: Any,
        color: str = "blue",
        highlight: Optional[List[Dict[str, str]]] = None,
    ) -> None:
        """Log a debug message."""
        if self.show_debugs:
            self._log("DEBUG", message, color, highlight, *args)

    def info(
        self,
        message: str,
        *args: Any,
        color: str = "green",
        highlight: Optional[List[Dict[str, str]]] = None,
    ) -> None:
        """Log an informational message."""
        self._log("INFO", message, color, highlight, *args)

    def warning(
        self,
        message: str,
        *args: Any,
        color: str = "yellow",
        highlight: Optional[List[Dict[str, str]]] = None,
    ) -> None:
        """Log a warning message."""
        self._log("WARNING", message, color, highlight, *args)

    def error(
        self,
        message: str,
        *args: Any,
        color: str = "red",
        highlight: Optional[List[Dict[str, str]]] = None,
    ) -> None:
        """Log an error message."""
        self._log("ERROR", message, color, highlight, *args)

    def critical(
        self,
        message: str,
        *args: Any,
        color: str = "magenta",
        highlight: Optional[List[Dict[str, str]]] = None,
    ) -> None:
        """Log a critical error message."""
        self._log("CRITICAL", message, color, highlight, *args)

    def _log(
        self,
        level: str,
        message: str,
        color: str,
        highlight: Optional[List[Dict[str, str]]],
        *args: Any,
    ) -> None:
        """Helper method to log messages with Rich formatting."""
        formatted_message = self.format_message(level, message, color, highlight)
        if self.progress is not None:
            # Print using console to ensure it does not interfere with progress bar
            self.console.print(formatted_message)
        else:
            # Use standard logging with rich formatting
            log_method = getattr(self.logger, level.lower())
            log_method(formatted_message, *args)

    def format_message(
        self,
        level: str,
        message: str,
        color: str,
        highlight: Optional[List[Dict[str, str]]],
    ) -> str:
        """Format the log message with appropriate colors based on the log level and highlight specific words."""
        formatted_message = f"[{color}]{level}: {message}[/{color}]"
        if highlight:
            for item in highlight:
                for word, highlight_color in item.items():
                    formatted_message = formatted_message.replace(
                        word, f"[{highlight_color}]{word}[/{highlight_color}]", 1
                    )
        return formatted_message

    def display_user_info(self, user_info: Dict[str, str]) -> None:
        """Display user information in a nicely formatted table."""
        table = Table(title="User Information")

        # Add columns
        table.add_column("Field", style="bold cyan")
        table.add_column("Value", style="bold white")

        # Add rows from user info dictionary
        for key, value in user_info.items():
            table.add_row(key, str(value))

        self.console.print(table)

    def start_progress_bar(self, name: str) -> None:
        """Start a progress bar for initialization."""
        self.progress = Progress(
            TextColumn("[progress.description]{task.description}"),
            BarColumn(),
            console=self.console,
        )

        self.task = self.progress.add_task(f"[blue]TASK: {name}...", total=100)
        self.progress.start()

    def update_progress_bar(self, advance: int = 10) -> None:
        """Update the progress bar by a certain amount."""
        if self.progress is not None and self.task is not None:
            self.progress.update(self.task, advance=advance)

    def stop_progress_bar(self) -> None:
        """Stop the progress bar."""
        if self.progress is not None and self.task is not None:
            self.progress.update(
                self.task, advance=self.progress.tasks[self.task].remaining
            )
            self.progress.stop()

    def inspect(self, obj: Any) -> None:
        """Inspect an object using the Rich library."""
        self.console.print(obj)
