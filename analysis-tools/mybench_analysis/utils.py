import numpy as np
from matplotlib.patches import Ellipse
from matplotlib.lines import Line2D
from matplotlib.text import Annotation


def data_to_ellipse(x, y, text=None, chisquare_scale = 2.77, **kwargs):
  # 2.77 is 75% of the data
  mu = np.mean((x, y), axis=1)
  covariance = np.cov((x, y), rowvar=True)
  eigenvals, eigenvecs = np.linalg.eig(covariance)

  max_ind = np.argmax(eigenvals)
  max_eigvec = eigenvecs[:,max_ind]
  max_eigval = eigenvals[max_ind]

  min_ind = 0
  if max_ind == 0:
    min_ind = 1

  min_eigvec = eigenvecs[:,min_ind]
  min_eigval = eigenvals[min_ind]

  width = 2 * np.sqrt(chisquare_scale * max_eigval)
  height = 2 * np.sqrt(chisquare_scale * min_eigval)
  angle = np.arctan2(max_eigvec[1], max_eigvec[0])

  # TODO: is this calculated correctly?
  min_x = mu[0] - width / 2 * np.cos(angle)
  max_x = mu[0] + width / 2 * np.cos(angle)

  min_y = mu[1] - width / 2 * np.sin(angle)
  max_y = mu[1] + width / 2 * np.sin(angle)

  return (
    Ellipse(xy=mu, width=width, height=height, angle=angle * 180.0 / np.pi, **kwargs),
    Line2D([min_x, max_x], [min_y, max_y], color=kwargs.get("edgecolor")),
    # rotation below is wrong, that's screen rotation, not angles, which means this only work when the aspect ratio is close to 1.
    None if text is None else Annotation(text, mu, xytext=(0, -2), textcoords="offset pixels", rotation=angle * 180.0 / np.pi, va="top", ha="center", color=kwargs.get("edgecolor"))
  )
