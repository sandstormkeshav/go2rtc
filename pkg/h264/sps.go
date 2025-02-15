package h264

import "github.com/AlexxIT/go2rtc/pkg/bits"

// http://www.itu.int/rec/T-REC-H.264
// https://webrtc.googlesource.com/src/+/refs/heads/main/common_video/h264/sps_parser.cc

//goland:noinspection GoSnakeCaseUsage
type SPS struct {
	profile_idc uint8
	profile_iop uint8
	level_idc   uint8

	seq_parameter_set_id uint32

	chroma_format_idc                    uint32
	separate_colour_plane_flag           byte
	bit_depth_luma_minus8                uint32
	bit_depth_chroma_minus8              uint32
	qpprime_y_zero_transform_bypass_flag byte
	seq_scaling_matrix_present_flag      byte

	log2_max_frame_num_minus4             uint32
	pic_order_cnt_type                    uint32
	log2_max_pic_order_cnt_lsb_minus4     uint32
	delta_pic_order_always_zero_flag      byte
	offset_for_non_ref_pic                int32
	offset_for_top_to_bottom_field        int32
	num_ref_frames_in_pic_order_cnt_cycle uint32
	num_ref_frames                        uint32
	gaps_in_frame_num_value_allowed_flag  byte

	pic_width_in_mbs_minus_1        uint32
	pic_height_in_map_units_minus_1 uint32
	frame_mbs_only_flag             byte
	mb_adaptive_frame_field_flag    byte
	direct_8x8_inference_flag       byte

	frame_cropping_flag      byte
	frame_crop_left_offset   uint32
	frame_crop_right_offset  uint32
	frame_crop_top_offset    uint32
	frame_crop_bottom_offset uint32

	vui_parameters_present_flag    byte
	aspect_ratio_info_present_flag byte
	aspect_ratio_idc               uint32
	sar_width                      uint32
	sar_height                     uint32
}

func (s *SPS) Width() uint16 {
	width := 16 * (s.pic_width_in_mbs_minus_1 + 1)
	crop := 2 * (s.frame_crop_left_offset + s.frame_crop_right_offset)
	return uint16(width - crop)
}

func (s *SPS) Height() uint16 {
	height := 16 * (s.pic_height_in_map_units_minus_1 + 1)
	crop := 2 * (s.frame_crop_top_offset + s.frame_crop_bottom_offset)
	if s.frame_mbs_only_flag == 0 {
		height *= 2
	}
	return uint16(height - crop)
}

func DecodeSPS(sps []byte) *SPS {
	r := bits.NewReader(sps)

	hdr := r.ReadByte()
	if hdr&0x1F != NALUTypeSPS {
		return nil
	}

	s := &SPS{
		profile_idc:          r.ReadByte(),
		profile_iop:          r.ReadByte(),
		level_idc:            r.ReadByte(),
		seq_parameter_set_id: r.ReadUEGolomb(),
	}

	switch s.profile_idc {
	case 100, 110, 122, 244, 44, 83, 86, 118, 128, 138, 139, 134, 135:
		n := byte(8)

		s.chroma_format_idc = r.ReadUEGolomb()
		if s.chroma_format_idc == 3 {
			s.separate_colour_plane_flag = r.ReadBit()
			n = 12
		}

		s.bit_depth_luma_minus8 = r.ReadUEGolomb()
		s.bit_depth_chroma_minus8 = r.ReadUEGolomb()
		s.qpprime_y_zero_transform_bypass_flag = r.ReadBit()

		s.seq_scaling_matrix_present_flag = r.ReadBit()
		if s.seq_scaling_matrix_present_flag != 0 {
			for i := byte(0); i < n; i++ {
				ssl := r.ReadBit() // seq_scaling_list_present_flag[i]
				if ssl != 0 {
					return nil // not implemented
				}
			}
		}
	}

	s.log2_max_frame_num_minus4 = r.ReadUEGolomb()

	s.pic_order_cnt_type = r.ReadUEGolomb()
	switch s.pic_order_cnt_type {
	case 0:
		s.log2_max_pic_order_cnt_lsb_minus4 = r.ReadUEGolomb()
	case 1:
		s.delta_pic_order_always_zero_flag = r.ReadBit()
		s.offset_for_non_ref_pic = r.ReadSEGolomb()
		s.offset_for_top_to_bottom_field = r.ReadSEGolomb()

		s.num_ref_frames_in_pic_order_cnt_cycle = r.ReadUEGolomb()
		for i := uint32(0); i < s.num_ref_frames_in_pic_order_cnt_cycle; i++ {
			_ = r.ReadSEGolomb() // offset_for_ref_frame[i]
		}
	}

	s.num_ref_frames = r.ReadUEGolomb()
	s.gaps_in_frame_num_value_allowed_flag = r.ReadBit()

	s.pic_width_in_mbs_minus_1 = r.ReadUEGolomb()
	s.pic_height_in_map_units_minus_1 = r.ReadUEGolomb()

	s.frame_mbs_only_flag = r.ReadBit()
	if s.frame_mbs_only_flag == 0 {
		s.mb_adaptive_frame_field_flag = r.ReadBit()
	}

	s.direct_8x8_inference_flag = r.ReadBit()

	s.frame_cropping_flag = r.ReadBit()
	if s.frame_cropping_flag != 0 {
		s.frame_crop_left_offset = r.ReadUEGolomb()
		s.frame_crop_right_offset = r.ReadUEGolomb()
		s.frame_crop_top_offset = r.ReadUEGolomb()
		s.frame_crop_bottom_offset = r.ReadUEGolomb()
	}

	s.vui_parameters_present_flag = r.ReadBit()
	if s.vui_parameters_present_flag != 0 {
		s.aspect_ratio_info_present_flag = r.ReadBit()
		if s.aspect_ratio_info_present_flag != 0 {
			s.aspect_ratio_idc = r.ReadBits(8)
			if s.aspect_ratio_idc == 255 {
				s.sar_width = r.ReadBits(16)
				s.sar_height = r.ReadBits(16)
			}
		}

		//...
	}

	if r.EOF {
		return nil
	}

	return s
}
