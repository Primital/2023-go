import matplotlib.pyplot as plt
import pandas as pd
import sys

# Provided data

if len(sys.argv) < 2:
    print("no file in argument")
    sys.exit(1)
filename = sys.argv[1]
try:
    df = pd.read_csv(filename)
except FileNotFoundError:
    print(f"File {filename} not found.")
    sys.exit(1)


# Create a new figure and axis for the plot
fig, ax1 = plt.subplots()

# Plot the score data on the left y-axis (logarithmic scale)
ax1.plot(df["generation"], df["best_solution"], label="Best Solution", linestyle='-')
ax1.plot(df["generation"], df["worst_solution"], label="Worst Solution", linestyle='-')
ax1.plot(df["generation"], df["average_score"], label="Average Score", linestyle='--')
ax1.set_yscale('log')  # Set the left y-axis to logarithmic scale

# Set labels and title for the left y-axis
ax1.set_xlabel("Generation")
ax1.set_ylabel("Score (log scale)", color='tab:blue')
ax1.tick_params(axis='y', labelcolor='tab:blue')
ax1.legend(loc='upper left')

# Create a second y-axis on the right side for diversity (0% to 100%)
ax2 = ax1.twinx()
ax2.plot(df["generation"], df["diversity"] * 100, label="Diversity (%)", color='tab:red', linestyle='-')
ax2.set_ylabel("Diversity (%)", color='tab:red')
ax2.tick_params(axis='y', labelcolor='tab:red')
ax2.set_ylim(0, 100)

# Set title and legend for the entire plot
plt.title("Progress of Genetic Algorithm Over Generations (Log Scale)")
plt.grid(True)
plt.legend(loc='upper right')

# Highlight points where "best_solution" changes/improves with 'X'
best_solution = df["best_solution"]
improvement_points = [best_solution.iloc[0]]  # Start with the initial value
generation_points = [df["generation"].iloc[0]]

for i in range(1, len(best_solution)):
    if best_solution.iloc[i] > best_solution.iloc[i - 1]:
        improvement_points.append(best_solution.iloc[i])
        generation_points.append(df["generation"].iloc[i])

ax1.scatter(generation_points, improvement_points, marker='o', color='blue', label='Improvement')
ax1.legend()

# Show the figure
plt.show()
# Plotting
#plt.figure(figsize=(10, 6))
#   plt.plot(df["generation"], df["best_solution"], label="Best Solution", linestyle='-')
#   plt.plot(df["generation"], df["worst_solution"], label="Worst Solution", linestyle='-')
#   plt.plot(df["generation"], df["average_score"], label="Average Score", linestyle='--')
#
#   #plt.yscale('log')  # Set the y-axis to logarithmic scale
#   plt.title("Progress of Genetic Algorithm Over Generations")
#   plt.xlabel("Generation")
#   plt.ylabel("Score")
#   plt.legend()
#   plt.grid(True)
#
#   # Create a new figure and axis for diversity gauge
#   fig, ax = plt.subplots()
#   ax.set_aspect('equal')  # Ensure the aspect ratio is equal for a circular gauge
#
#   # Plot a filled circle with color indicating diversity (0.0 to 1.0)
#   diversity = 0.8  # Replace with your actual diversity value
#   circle = plt.Circle((0.5, 0.5), 0.4, color=plt.cm.viridis(diversity), fill=True)
#   ax.add_artist(circle)
#
#   # Set the title and remove axis labels for the diversity gauge
#   ax.set_title("Diversity Gauge")
#   ax.set_xticks([])
#   ax.set_yticks([])
#
#   # Show the figure
#   plt.show()
