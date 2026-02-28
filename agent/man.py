import itertools
import time


BASE_ART = [
    "                                                                    ...",
    "                                                                  .##-.",
    "                                           .. ..... .....         .##-.",
    "                                             ..-)}}(+:.=~.       .<##-.",
    "                                            ..>{{{####{#(.   .. .=>##-.",
    "                                            ..##{#{><]{[:..   ..:>*##-.",
    "                                            ..{#**<****.... ....:>*##-.",
    "                                           ...}(******=.....-*>***>##-.",
    "                                    .......::+**<****~.:*>*****>+..##-.",
    "                                     .:>^^<<>^^^^^<*********>^:....##-.",
    "                                 ....)^^^^^^^^^^^*<******>>~.......##-.",
    "                                 ...<^^^^^^^^^^^^*)<<^=-:..........##-.",
    "                             .....:>^^^^^^<<<<<>^~......  ..   ....##-.",
    "                             .....>^^^^^^^^^^^^^: ..             ..##-.",
    "                              ...<^^^^^^^^^^^^^+. .              ..##-.",
    "                             ...)^^^^^^^^^^^^<-.                 ..##-.",
    "                              .<^^^^^^^^^^<>*. ....              ..##-.",
    "                           ..~>^^^^^^^<>^^<*.......              ..##-.",
    "                          ..>^^^^^^^<))^^^^:.......              ..##-.",
    "                          .:#{#{]<^^^^^^^^~....                  ..##-.",
    "                          .]{#########{[(^..   .                 ..##-.",
    "                          .#############}{{+.  ....              ..##-.",
    "                          .#{##########]#####+.....              ..##-.",
    "                         .)#{########{]{#######=...              ..##-.",
    "                     ....-{#########{[##########{:.              ..##-.",
    "                     ....}#########{+{###########{]..            ..##-.",
    "                      ..>########{)....=}{#########}...          ..##-.",
    "                      .~########{=........:)#{#####{...          ..##-.",
    "                  ....-########}.           .[#####]...          ..##-.",
    "                   .(#########=..            {###{{^...          ..##-.",
    "               ...}{{{{{{###^...            .######:.            ..##-.",
    "               .]###{{{{{}:...              .%####[.             ..##-.",
    "           ...*####{{##^.....               -####{=              ..##-.",
    "          ..:#######]. .                   .^{####.              ..##-.",
    "       . ..>{{{{{#=...                     .[###{+..  .          ..##-.",
    "       .-}{**{#<. .....                    .{{###. .  .          ..##-.",
    "      .:#{###:.                            .#{{{+...  .          ..##-.",
    " ...  ...>###}.....          ....        ..:}(]{<.... ....   ......##-.",
    "        ..:{{{}<:...         ....        ..(#{#{#{#{)-...      ....##-.",
    "::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::",
]

OFFSETS = [0, 1, 2, 3, 2, 1, 0, -1, -2, -1, 0, 1]


def _shift_region(text: str, offset: int, width: int) -> str:
    if offset > 0:
        shifted = (" " * offset) + text
        return shifted[:width]
    if offset < 0:
        shifted = text[-offset:]
        return shifted.ljust(width)
    return text[:width].ljust(width)


def _build_frame(lines: list[str], offset: int) -> str:
    max_width = max(len(line) for line in lines)
    out = []
    floor_idx = len(lines) - 1
    for idx, line in enumerate(lines):
        padded = line.ljust(max_width)
        if idx == floor_idx:
            out.append(padded)
            continue
        out.append(_shift_region(padded, offset, max_width))
    return "\n".join(out)


def build_frames() -> list[str]:
    return [_build_frame(BASE_ART, offset) for offset in OFFSETS]


def render_static() -> str:
    return build_frames()[0]


def animate(repeat_delay: float = 0.08) -> None:
    frames = build_frames()
    frame_height = len(BASE_ART)
    move_up = max(frame_height - 1, 0)
    first = True
    try:
        for frame in itertools.cycle(frames):
            if first:
                # Initial paint reserves the region in terminal scrollback.
                print(frame, end="", flush=True)
                first = False
            else:
                # Move cursor back to frame start and repaint in place.
                print(f"\033[{move_up}F\033[J{frame}", end="", flush=True)
            time.sleep(repeat_delay)
    finally:
        # Leave cursor below the animation block on exit.
        print()
