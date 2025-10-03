
/* RescueTime::Network::API::perform_activate(std::__cxx11::basic_string<wchar_t,
   std::char_traits<wchar_t>, std::allocator<wchar_t> > const&, std::__cxx11::basic_string<wchar_t,
   std::char_traits<wchar_t>, std::allocator<wchar_t> > const&, std::__cxx11::basic_string<wchar_t,
   std::char_traits<wchar_t>, std::allocator<wchar_t> > const&, std::__cxx11::basic_string<wchar_t,
   std::char_traits<wchar_t>, std::allocator<wchar_t> > const&) */

basic_string *
RescueTime::Network::API::perform_activate
          (basic_string *param_1,basic_string *param_2,basic_string *param_3,basic_string *param_4)

{
  ulong uVar1;
  int iVar2;
  basic_string *pbVar3;
  basic_string *pbVar4;
  ui *this;
  basic_json *pbVar5;
  size_t sVar6;
  _Rb_tree<> *p_Var7;
  _Rb_tree<> *p_Var8;
  utf8 *this_00;
  basic_string *in_R8;
  basic_string *in_R9;
  utf8 *this_01;
  long in_FS_OFFSET;
  utf8 local_2b8 [16];
  utf8 local_2a8 [16];
  utf8 local_298 [16];
  utf8 local_288 [16];
  utf8 local_278 [16];
  utf8 local_268 [16];
  utf8 local_258 [16];
  utf8 local_248 [16];
  utf8 local_238 [16];
  utf8 local_228 [16];
  byte local_218;
  _Rb_tree<> *local_210;
  input_adapter local_208 [8];
  long local_200;
  basic_json<> local_1f8 [8];
  json_value local_1f0 [8];
  undefined *local_1e8 [2];
  undefined auStack_1d8 [16];
  wchar_t *local_1c8;
  int local_1c0;
  wchar_t awStack_1b8 [4];
  undefined *local_1a8;
  undefined8 local_1a0;
  undefined local_198 [16];
  undefined *local_188;
  undefined8 local_180;
  undefined local_178 [16];
  undefined *local_168 [2];
  undefined auStack_158 [16];
  wchar_t *local_148;
  ulong local_140;
  wchar_t local_138 [4];
  code **local_128 [2];
  code *local_118 [2];
  undefined *local_108;
  long local_100;
  undefined auStack_f8 [24];
  _Rb_tree<> local_e0 [16];
  _Rb_tree_node *local_d0;
  request local_a8 [104];
  long local_40;

  local_40 = *(long *)(in_FS_OFFSET + 0x28);
                    /* try { // try from 0063066f to 00630673 has its CatchHandler @ 0063108e */
  std::__cxx11::basic_string<>::basic_string
            ((wchar_t *)local_1e8,(allocator *)L"[API::perform_activate]");
                    /* try { // try from 0063067c to 00630680 has its CatchHandler @ 00631068 */
  rt::debug_log((basic_string *)local_1e8,0xc0);
  if (local_1e8[0] != auStack_1d8) {
    operator.delete(local_1e8[0]);
  }
                    /* try { // try from 00630696 to 006306d2 has its CatchHandler @ 0063108e */
  rt::config::get_instance();
  pbVar3 = (basic_string *)rt::config::get_account_key[abi:cxx11]();
  rt::config::get_instance();
  pbVar4 = (basic_string *)rt::config::get_url[abi:cxx11]();
  std::__cxx11::basic_string<>::basic_string((char *)&local_1c8,(allocator *)"/activate");
                    /* try { // try from 006306ed to 006306f1 has its CatchHandler @ 00630fa4 */
  rt::network::request::request(local_a8,(basic_string *)&local_1c8,pbVar4,pbVar3);
  if (local_1c8 != awStack_1b8) {
    operator.delete(local_1c8);
  }
                    /* try { // try from 0063071b to 0063071f has its CatchHandler @ 00630f99 */
  rt::utf8::utf8(local_2a8,"application/json");
                    /* try { // try from 00630731 to 00630735 has its CatchHandler @ 00630f97 */
  rt::utf8::utf8(local_2b8,"Accept");
                    /* try { // try from 0063073f to 00630743 has its CatchHandler @ 00630f92 */
  rt::network::request::add_header((utf8 *)local_a8,local_2b8);
  rt::utf8::~utf8(local_2b8);
  rt::utf8::~utf8(local_2a8);
                    /* try { // try from 00630754 to 0063078c has its CatchHandler @ 00630f99 */
  this = (ui *)rt::ui_factory::get_ui();
  rt::ui::add_activate_params(this,local_a8);
  iVar2 = std::__cxx11::basic_string<>::compare((basic_string<> *)in_R9,L"");
  if (iVar2 == 0) {
                    /* try { // try from 00630a5e to 00630a9b has its CatchHandler @ 00630f99 */
    iVar2 = std::__cxx11::basic_string<>::compare((basic_string<> *)param_3,L"");
    if ((iVar2 != 0) &&
       (iVar2 = std::__cxx11::basic_string<>::compare((basic_string<> *)param_4,L""), iVar2 != 0)) {
      rt::utf8::utf8(local_268,param_3);
                    /* try { // try from 00630aad to 00630ab1 has its CatchHandler @ 00630f4c */
      rt::utf8::utf8(local_278,"username");
                    /* try { // try from 00630abb to 00630abf has its CatchHandler @ 00630f2b */
      rt::network::request::add_param((utf8 *)local_a8,local_278);
      rt::utf8::~utf8(local_278);
      rt::utf8::~utf8(local_268);
                    /* try { // try from 00630ae1 to 00630ae5 has its CatchHandler @ 00630f99 */
      rt::utf8::utf8(local_248,param_4);
                    /* try { // try from 00630af7 to 00630afb has its CatchHandler @ 00630f23 */
      rt::utf8::utf8(local_258,"password");
                    /* try { // try from 00630b05 to 00630b09 has its CatchHandler @ 00630f02 */
      rt::network::request::add_param((utf8 *)local_a8,local_258);
      rt::utf8::~utf8(local_258);
      rt::utf8::~utf8(local_248);
                    /* try { // try from 00630b28 to 00630b4a has its CatchHandler @ 00630f99 */
      iVar2 = std::__cxx11::basic_string<>::compare((basic_string<> *)in_R8,L"");
      if (iVar2 != 0) {
        this_00 = local_228;
        rt::utf8::utf8(this_00,in_R8);
        this_01 = local_238;
                    /* try { // try from 00630b5c to 00630b60 has its CatchHandler @ 006310ac */
        rt::utf8::utf8(this_01,"two_factor_auth_code");
                    /* try { // try from 00630b6a to 00630b6e has its CatchHandler @ 006310a7 */
        rt::network::request::add_param((utf8 *)local_a8,this_01);
        goto LAB_006307b1;
      }
    }
  }
  else {
    this_00 = local_288;
    rt::utf8::utf8(this_00,in_R9);
    this_01 = local_298;
                    /* try { // try from 0063079e to 006307a2 has its CatchHandler @ 00630f90 */
    rt::utf8::utf8(this_01,"enterprise_team_key");
                    /* try { // try from 006307ac to 006307b0 has its CatchHandler @ 00630f8b */
    rt::network::request::add_param((utf8 *)local_a8,this_01);
LAB_006307b1:
    rt::utf8::~utf8(this_01);
    rt::utf8::~utf8(this_00);
  }
                    /* try { // try from 006307d3 to 006307d7 has its CatchHandler @ 00630f99 */
  rt::network::request::perform((request *)&local_108,(method)local_a8);
                    /* try { // try from 006307e0 to 006307e4 has its CatchHandler @ 00630f80 */
  validate_response_status_code((Response *)&local_108,200);
  local_118[0] = (code *)0x0;
                    /* try { // try from 0063080b to 0063080f has its CatchHandler @ 00630f54 */
  nlohmann::detail::input_adapter::input_adapter<>(local_208,local_108,local_108 + local_100);
                    /* try { // try from 00630833 to 00630837 has its CatchHandler @ 0063103b */
  nlohmann::basic_json<>::parse((basic_json<> *)&local_218,local_208,local_128,1);
  if (local_200 != 0) {
    std::_Sp_counted_base<>::_M_release();
  }
  if (local_118[0] != (code *)0x0) {
    (*local_118[0])(local_128,local_128,3);
  }
                    /* try { // try from 00630873 to 00630877 has its CatchHandler @ 00631030 */
  nlohmann::basic_json<>::basic_json(local_1f8,(basic_json *)&local_218);
                    /* try { // try from 0063087b to 0063087f has its CatchHandler @ 00630fff */
  throw_exception_if_rt_error_present((basic_json)local_1f8);
  nlohmann::basic_json<>::assert_invariant();
  nlohmann::basic_json<>::json_value::destroy(local_1f0,(uint)(byte)local_1f8[0]);
                    /* try { // try from 006308a2 to 006308a6 has its CatchHandler @ 00631030 */
  pbVar5 = nlohmann::basic_json<>::operator[]<char_const>((basic_json<> *)&local_218,"account_key");
  local_1a0 = 0;
  local_198[0] = 0;
  local_1a8 = local_198;
                    /* try { // try from 006308d1 to 006308d5 has its CatchHandler @ 00630fd1 */
  nlohmann::detail::from_json<>(pbVar5,(string_t *)&local_1a8);
                    /* try { // try from 006308e0 to 006308e4 has its CatchHandler @ 0063105b */
  rt::utf8_to_wstring((rt *)&local_1c8,(basic_string *)&local_1a8);
  if (local_1a8 != local_198) {
    operator.delete(local_1a8);
  }
  local_140 = 0;
  local_138[0] = L'\0';
  local_148 = local_138;
  if (local_218 == 1) {
    p_Var7 = local_210 + 8;
                    /* try { // try from 00630b98 to 00630bf5 has its CatchHandler @ 00631060 */
    std::__cxx11::basic_string<>::basic_string((char *)local_128,(allocator *)"data_key");
    p_Var8 = (_Rb_tree<> *)std::_Rb_tree<>::find(local_210,(basic_string *)local_128);
    if (local_128[0] != local_118) {
      operator.delete(local_128[0]);
    }
    if (p_Var7 != p_Var8) {
      pbVar5 = nlohmann::basic_json<>::operator[]<char_const>((basic_json<> *)&local_218,"data_key")
      ;
      local_180 = 0;
      local_178[0] = 0;
      local_188 = local_178;
                    /* try { // try from 00630c27 to 00630c2b has its CatchHandler @ 00631096 */
      nlohmann::detail::from_json<>(pbVar5,(string_t *)&local_188);
                    /* try { // try from 00630c44 to 00630c48 has its CatchHandler @ 00630c9e */
      rt::utf8_to_wstring((rt *)local_168,(basic_string *)&local_188);
      std::__cxx11::basic_string<>::swap((basic_string<> *)&local_148,(basic_string *)local_168);
      if (local_168[0] != auStack_158) {
        operator.delete(local_168[0]);
      }
      if (local_188 != local_178) {
        operator.delete(local_188);
      }
      goto LAB_0063095e;
    }
  }
  uVar1 = local_140;
  sVar6 = wcslen(L"");
                    /* try { // try from 00630959 to 006309b0 has its CatchHandler @ 00631060 */
  std::__cxx11::basic_string<>::_M_replace((basic_string<> *)&local_148,0,uVar1,L"",sVar6);
LAB_0063095e:
  *(basic_string **)param_1 = param_1 + 4;
  std::__cxx11::basic_string<>::_M_construct<>
            ((wchar_t *)param_1,local_1c8,(int)local_1c8 + local_1c0 * 4);
  *(basic_string **)(param_1 + 8) = param_1 + 0xc;
  std::__cxx11::basic_string<>::_M_construct<>
            ((wchar_t *)(param_1 + 8),local_148,(int)local_148 + (int)local_140 * 4);
  if (local_148 != local_138) {
    operator.delete(local_148);
  }
  if (local_1c8 != awStack_1b8) {
    operator.delete(local_1c8);
  }
  nlohmann::basic_json<>::assert_invariant();
  nlohmann::basic_json<>::json_value::destroy((json_value *)&local_210,(uint)local_218);
  std::_Rb_tree<>::_M_erase(local_e0,local_d0);
  if (local_108 != auStack_f8) {
    operator.delete(local_108);
  }
  rt::network::request::~request(local_a8);
  if (local_40 != *(long *)(in_FS_OFFSET + 0x28)) {
                    /* WARNING: Subroutine does not return */
    __stack_chk_fail();
  }
  return param_1;
}
