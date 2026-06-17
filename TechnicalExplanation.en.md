🌐 **English** | [日本語](TechnicalExplanation.ja.md)
# Technical Explanation

The Vectrace process is not merely an extension of binarization (black-and-white conversion) but consists of four distinct phases: "Color Clustering," "Layer Separation," "Shape Normalization (Morphology)," and "Vectorization."

## 1. Color Analysis and Clustering

First, the vast number of colors in the image is aggregated into a manageable set (up to 256 colors).

- **Pixel Scanning**: Every pixel is inspected, excluding alpha values (transparency) and pure white (often treated as background).
- **Perceptual Distance Integration**: Using the `colorDistance` function, the proximity of colors is determined based on human visual sensitivity (Luma weights: R=29.9%, G=58.7%, B=11.4%).
- **Mean Color Calculation**: Similar colors are grouped into "clusters," and the average RGB value of all pixels in each group is determined as the representative color for that layer. This maintains the nuance of the original image.

## 2. Layer Separation and Base Layer Extraction

A "bitmap" (filled area information) is created for each extracted color.

- **Color Layers**: Each pixel is assigned to the bitmap of the representative color with the shortest perceptual distance.
- **Base Layer (Outline)**: A special layer is created to extract dark areas based on the `-k` (threshold) flag. It uses the average color of the extracted range (or forced to black with `-K`) to ensure smooth blending at boundaries.
- **Background Silhouette**: A "white base" covering all opaque regions is created and placed at the very back of the stack.

## 3. Image Enhancement via Morphology

Raw bitmap data often contains "isolated points" and "jaggies." The shapes are refined before tracing.

- **Dilation and Erosion**: Functions like `dilateBitmap` and `erodeBitmap` are executed to fill gaps and suppress sharp noise (aliasing).
- **Adaptive Pass Count**: The number of iterations is dynamically adjusted based on the area ratio of the layer (larger areas receive stronger smoothing). This protects fine details while removing noise from large filled surfaces.
- **Boundary Adjustment**: If the `-b` (bg-dilation) flag is specified, additional dilation is applied to the background silhouette to ensure a margin that hides gaps between colors.

## 4. Parallel Tracing and SVG Compositing

Finally, each bitmap is converted into vector paths and layered together.

- **Parallel Processing**: Using `errgroup`, the tracing of all layers is executed in parallel, fully utilizing CPU threads.
- **Z-Order Control**:
  1. **Bottom**: White background silhouette.
  2. **Middle**: Base layer (outline).
  3. **Top**: Color layers **rendered in order of descending area** (placing smaller parts on top of larger fills to maintain detail).
- **SVG Assembly**: Each layer is fragmented into SVG paths via `vectrace.Render` and then combined into a single `<svg>` tag within a group (`<g>`).

By combining "statistical color aggregation" with "layer-specific shape correction," Vectrace converts complex color images into clean vector data with minimal artifacts.
