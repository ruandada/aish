#!/usr/bin/env python3
"""
🎨 AI Plotting Tool Demo Script
Demonstrates how to generate various types of charts using Python
"""

import argparse
import matplotlib.pyplot as plt
import numpy as np
from matplotlib import rcParams

# Set font configuration for better display
rcParams['font.sans-serif'] = ['Arial', 'DejaVu Sans', 'Liberation Sans']
rcParams['axes.unicode_minus'] = False

def print_banner():
    """Print tool banner"""
    banner = """
╔══════════════════════════════════════════════════════════════╗
║                 🎨 AI Plotting Tool Demo 🎨                 ║
║                  Python Matplotlib Chart Generator           ║
╚══════════════════════════════════════════════════════════════╝
    """
    print(banner)

def create_line_chart(title="Line Chart"):
    """Create line chart"""
    x = np.linspace(0, 10, 100)
    y1 = np.sin(x)
    y2 = np.cos(x)
    
    plt.figure(figsize=(10, 6))
    plt.plot(x, y1, label='sin(x)', color='blue', linewidth=2)
    plt.plot(x, y2, label='cos(x)', color='red', linewidth=2)
    plt.title(title, fontsize=16, fontweight='bold')
    plt.xlabel('X Axis', fontsize=12)
    plt.ylabel('Y Axis', fontsize=12)
    plt.legend()
    plt.grid(True, alpha=0.3)
    plt.show()
    print(f"✅ Line chart displayed: {title}")

def create_bar_chart(title="Bar Chart"):
    """Create bar chart"""
    categories = ['Apple', 'Banana', 'Orange', 'Grape', 'Strawberry']
    values = [23, 45, 56, 78, 32]
    
    plt.figure(figsize=(10, 6))
    bars = plt.bar(categories, values, color=['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7'])
    plt.title(title, fontsize=16, fontweight='bold')
    plt.xlabel('Fruit Types', fontsize=12)
    plt.ylabel('Quantity', fontsize=12)
    
    # Add value labels on bars
    for bar, value in zip(bars, values):
        plt.text(bar.get_x() + bar.get_width()/2, bar.get_height() + 1, 
                str(value), ha='center', va='bottom', fontweight='bold')
    
    plt.show()
    print(f"✅ Bar chart displayed: {title}")

def create_pie_chart(title="Pie Chart"):
    """Create pie chart"""
    labels = ['Technology', 'Design', 'Marketing', 'Operations', 'Other']
    sizes = [30, 25, 20, 15, 10]
    colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7']
    
    plt.figure(figsize=(8, 8))
    plt.pie(sizes, labels=labels, colors=colors, autopct='%1.1f%%', 
            startangle=90, shadow=True)
    plt.title(title, fontsize=16, fontweight='bold')
    plt.axis('equal')
    plt.show()
    print(f"✅ Pie chart displayed: {title}")

def create_scatter_plot(title="Scatter Plot"):
    """Create scatter plot"""
    np.random.seed(42)
    x = np.random.randn(100)
    y = 2 * x + np.random.randn(100) * 0.5
    
    plt.figure(figsize=(10, 6))
    plt.scatter(x, y, alpha=0.6, c='purple', s=50)
    plt.title(title, fontsize=16, fontweight='bold')
    plt.xlabel('X Axis', fontsize=12)
    plt.ylabel('Y Axis', fontsize=12)
    plt.grid(True, alpha=0.3)
    plt.show()
    print(f"✅ Scatter plot displayed: {title}")

def create_heatmap(title="Heatmap"):
    """Create heatmap"""
    data = np.random.rand(10, 10)
    
    plt.figure(figsize=(8, 6))
    plt.imshow(data, cmap='viridis', aspect='auto')
    plt.colorbar(label='Value')
    plt.title(title, fontsize=16, fontweight='bold')
    plt.xlabel('X Axis', fontsize=12)
    plt.ylabel('Y Axis', fontsize=12)
    plt.show()
    print(f"✅ Heatmap displayed: {title}")

def main():
    """Main function"""
    parser = argparse.ArgumentParser(description='AI Plotting Tool Demo')
    parser.add_argument('chart_type', choices=['line', 'bar', 'pie', 'scatter', 'heatmap', 'all'],
                       help='Chart type')
    parser.add_argument('--title', default='', help='Chart title')
    
    args = parser.parse_args()
    
    print_banner()
    
    if args.chart_type == 'all':
        # Generate all types of charts
        charts = [
            ('line', 'Line Chart'),
            ('bar', 'Bar Chart'),
            ('pie', 'Pie Chart'),
            ('scatter', 'Scatter Plot'),
            ('heatmap', 'Heatmap')
        ]
        
        for chart_type, title in charts:
            if args.title:
                title = f"{args.title} - {title}"
            
            if chart_type == 'line':
                create_line_chart(title)
            elif chart_type == 'bar':
                create_bar_chart(title)
            elif chart_type == 'pie':
                create_pie_chart(title)
            elif chart_type == 'scatter':
                create_scatter_plot(title)
            elif chart_type == 'heatmap':
                create_heatmap(title)
    else:
        # Generate single chart
        title = args.title or f"{args.chart_type.title()} Chart"
        
        if args.chart_type == 'line':
            create_line_chart(title)
        elif args.chart_type == 'bar':
            create_bar_chart(title)
        elif args.chart_type == 'pie':
            create_pie_chart(title)
        elif args.chart_type == 'scatter':
            create_scatter_plot(title)
        elif args.chart_type == 'heatmap':
            create_heatmap(title)

if __name__ == "__main__":
    main() 